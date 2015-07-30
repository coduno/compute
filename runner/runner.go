package runner

import (
	"io/ioutil"
	"net/http"
	"path"

	"github.com/coduno/compute/docker"
)

var (
	fileNames = map[string]string{
		"py":   "app.py",
		"c":    "app.c",
		"cpp":  "app.cpp",
		"java": "Application.java",
	}
)

type CodeTask struct {
	Flags,
	Code,
	Runner,
	Language string
}

// RunHandler is the general handler for the whole workflow.
// The base setupRunHandler is:
// - rh.Handle
// - start docker run
// - rh.Respond
type RunHandler interface {
	Handle(task CodeTask, w http.ResponseWriter, r *http.Request) docker.Config
	Respond(w http.ResponseWriter, req *http.Request, res docker.Result)
}

// CodeData is the data to receive from the codeground
type CodeData struct {
	CodeBase string `json:"codeBase"`
	Token    string `json:"token"`
	Language string `json:"language"`
}

// GeneralHandle function for a simple run. It writes the file with code in
// the tmp folder and  returns the docker run configuration.
func GeneralHandle(task CodeTask, w http.ResponseWriter, r *http.Request) (c docker.Config) {
	// TODO(victorbalan): POST Method check

	for availableLanguage := range fileNames {
		if task.Language == availableLanguage {
			goto LANGUAGE_AVAILABLE
		}
	}
	http.Error(w, "language not available", http.StatusBadRequest)
	return

LANGUAGE_AVAILABLE:
	c, err := docker.NewConfig(docker.NewImage(task.Language), "", task.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileName := path.Join(c.Volume, fileNames[task.Language])
	if err := ioutil.WriteFile(fileName, []byte(task.Code), 0777); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return
}
