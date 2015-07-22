package runner

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/coduno/app/models"
	"github.com/coduno/piper/docker"
	"github.com/coduno/piper/logger"
	"github.com/coduno/piper/piper"
)

// RunResults holds the results from a general run and the tmpDir
type RunResults struct {
	tmpDir string
	runOut string
	runErr string
}

// Run is the interface that GeneralRun uses to handle the output
type Run interface {
	computeResults(w http.ResponseWriter, runResults RunResults)
}

// GeneralRun represents the general part of a docker run. It computes
// the results via the run parameter
func GeneralRun(w http.ResponseWriter, r *http.Request, tmpDir string, codeData *models.CodeData, run Run) {
	key, build := logger.LogBuildStart("challengeId", codeData.CodeBase, "user")

	volume, err := docker.Dockerize(tmpDir)

	if err != nil {
		log.Fatal(err)
	}
	cmdUser := docker.NewDockerCmd(codeData.Language, volume)
	var runOut, runErr bytes.Buffer

	outUser, errUser, _, err := docker.GetStreams(&cmdUser.Cmd)
	if err != nil {
		return
	}

	cmdUser.Start()
	var wg sync.WaitGroup
	wg.Add(2)

	go piper.PipeOutput(&wg, outUser, os.Stdout, &runOut)
	go piper.PipeOutput(&wg, errUser, os.Stdout, &runErr)

	exitErr := cmdUser.Wait()
	wg.Wait()

	prepLog, stats := logger.GetLogs(tmpDir)
	logger.LogRunComplete(key, build, "", runOut.String(), "", exitErr, prepLog, stats)

	run.computeResults(w, RunResults{tmpDir, runOut.String(), runErr.String()})
}
