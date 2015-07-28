package runner

import (
	"encoding/json"
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

// RunHandler is the general handler for the whole workflow.
// The base setupRunHandler is:
// - rh.Handle
// - start docker run
// - rh.Respond
type RunHandler interface {
	Handle(w http.ResponseWriter, r *http.Request) docker.Config
	Respond(w http.ResponseWriter, req *http.Request, res docker.Result)
}

// CodeData is the data to receive from the codeground
type CodeData struct {
	CodeBase string `json:"codeBase"`
	Token    string `json:"token"`
	Language string `json:"language"`
}

func getCodeDataFromRequest(r *http.Request) (codeData CodeData, err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &codeData)
	return
}

// GeneralHandle function for a simple run. It writes the file with code in
// the tmp folder and  returns the docker run configuration.
func GeneralHandle(w http.ResponseWriter, r *http.Request) (c docker.Config) {
	// TODO(victorbalan): POST Method check

	codeData, err := getCodeDataFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	for availableLanguage := range fileNames {
		if codeData.Language == availableLanguage {
			goto LANGUAGE_AVAILABLE
		}
	}
	http.Error(w, "language not available", http.StatusBadRequest)
	return

LANGUAGE_AVAILABLE:
	c, err = docker.NewConfig(docker.NewImage(codeData.Language), "", codeData.CodeBase)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = ioutil.WriteFile(path.Join(c.Volume, fileNames[codeData.Language]), []byte(codeData.CodeBase), 0777)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}
