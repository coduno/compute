package runner

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"sync"

	"github.com/coduno/app/models"
	"github.com/coduno/piper/docker"
	"github.com/coduno/piper/piper"
)

var (
	fileNames = map[string]string{
		"py":   "app.py",
		"c":    "app.c",
		"cpp":  "app.cpp",
		"java": "Application.java",
	}
)

// RunHandler is the general handler for the whole workflow.
// The base setupRunHandler is:
// - rh.Handle
// - start docker run
// - rh.Respond
type RunHandler interface {
	Handle(r *http.Request) (docker.DockerConfig, error)
	Respond(w http.ResponseWriter, r *http.Request, rr RunResults)
}

// RunResults holds the results from a general run and the tmpDir
type RunResults struct {
	RunOut string
	RunErr string
}

// GeneralRun represents the general part of a docker run. It returns the run
// results
func GeneralRun(w http.ResponseWriter, r *http.Request, dc docker.DockerConfig) (rr RunResults) {
	cmdUser := docker.NewDockerCmd(dc.Image, dc.Volume)
	var runOut, runErr bytes.Buffer

	outUser, errUser, _, err := docker.GetStreams(cmdUser)
	if err != nil {
		return
	}

	cmdUser.Start()
	var wg sync.WaitGroup
	wg.Add(2)

	go piper.PipeOutput(&wg, outUser, os.Stdout, &runOut)
	go piper.PipeOutput(&wg, errUser, os.Stdout, &runErr)

	cmdUser.Wait()
	wg.Wait()
	rr.RunErr = runErr.String()
	rr.RunOut = runOut.String()
	return
}

func getCodeDataFromRequest(r *http.Request) (codeData models.CodeData, err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &codeData)
	return
}
