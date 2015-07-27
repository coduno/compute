package runner

import (
	"encoding/json"
	"net/http"

	"github.com/coduno/compute/docker"
)

// SimpleRunHandler is the handler for a simple run
type SimpleRunHandler struct {
}

// Handle function for a simple run. It writes the file with code in
// the tmp folder and  returns the docker run configuration.
func (srh SimpleRunHandler) Handle(w http.ResponseWriter, r *http.Request) (c docker.Config) {
	return GeneralHandle(w, r)
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
