package runner

import (
	"encoding/json"
	"net/http"

	"github.com/coduno/app/models"
)

// RunSimple is the struct to be passed to the GeneralRun function
type RunSimple struct {
}

func (s RunSimple) computeResults(w http.ResponseWriter, runResults RunResults) {

	var toSend = make(map[string]interface{})
	toSend["run"] = runResults.runOut
	toSend["err"] = runResults.runErr
	json, err := json.Marshal(toSend)
	if err != nil {
		http.Error(w, "Json marshal err: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(json)
}

// SimpleRun describes a simple run with just the run output and the errors
// sent back to the client
func SimpleRun(w http.ResponseWriter, r *http.Request, tmpDir string, codeData *models.CodeData) {
	var run RunSimple
	GeneralRun(w, r, tmpDir, codeData, run)
}
