package runner

import (
	"encoding/json"
	"net/http"
)

// SimpleRun is the struct to be passed to the GeneralRun function
type SimpleRun struct {
}

func (s SimpleRun) computeResults(w http.ResponseWriter, runResults RunResults) {

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
