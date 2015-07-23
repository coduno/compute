package runner

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/coduno/piper/docker"
)

// SimpleRunHandler is the handler for a simple run
type SimpleRunHandler struct {
}

// Handle function for a simple run. It writes the file with code in
//the tmp folder and  returns the docker run configuration
func (srh SimpleRunHandler) Handle(r *http.Request) (dc docker.DockerConfig, err error) {
	// TODO(victorbalan): POST Method check

	codeData, err := getCodeDataFromRequest(r)
	if err != nil {
		return
	}

	for availableLanguage := range fileNames {
		if codeData.Language == availableLanguage {
			goto LANGUAGE_AVAILABLE
		}
	}
	err = errors.New("Language not available.")
	return

LANGUAGE_AVAILABLE:
	tmpDir, err := docker.VolumeDir()
	if err != nil {
		return
	}
	err = ioutil.WriteFile(path.Join(tmpDir, fileNames[codeData.Language]), []byte(codeData.CodeBase), 0777)
	if err != nil {
		return
	}
	volume, err := docker.Dockerize(tmpDir)
	if err != nil {
		return
	}
	dc.TmpDir = tmpDir
	dc.Volume = volume
	dc.Image = docker.GetImageForLanguage(codeData.Language)
	return
}

// Respond implementation for a simple run. It returns the run output and the
// error output.
func (srh SimpleRunHandler) Respond(w http.ResponseWriter, r *http.Request, rr RunResults) {
	var toSend = make(map[string]interface{})
	toSend["run"] = rr.RunOut
	toSend["err"] = rr.RunErr
	json, err := json.Marshal(toSend)
	if err != nil {
		http.Error(w, "Json marshal err: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(json)
}
