package runner

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/coduno/piper/docker"
)

// SimpleRunHandler is the handler for a simple run
type SimpleRunHandler struct{}

// Handle function for a simple run. It writes the file with code in
// the tmp folder and  returns the docker run configuration.
func (srh SimpleRunHandler) Handle(w http.ResponseWriter, r *http.Request) (c docker.Config) {
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
	c, err = docker.NewConfig(docker.NewImage(codeData.Language), "")
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

// Respond implementation for a simple run. It returns the run output and the
// error output.
func (srh SimpleRunHandler) Respond(w http.ResponseWriter, req *http.Request, res docker.Result) {
	json, err := json.Marshal(map[string]string{
		"stdout": string(res.Stdout.Bytes()),
		"stderr": string(res.Stderr.Bytes()),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(json)
}
