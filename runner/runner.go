package runner

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/coduno/app/models"
	"github.com/coduno/piper/docker"
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
	Handle(w http.ResponseWriter, r *http.Request) docker.Config
	Respond(w http.ResponseWriter, req *http.Request, res docker.Result)
}

func getCodeDataFromRequest(r *http.Request) (codeData models.CodeData, err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &codeData)
	return
}
