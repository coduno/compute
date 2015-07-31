package runner

import (
	"encoding/json"
	"net/http"
)

// Handle function for a simple run. It writes the file with code in
// the tmp folder and  returns the docker run configuration.
func Simple(w http.ResponseWriter, r *http.Request) {
	task := decode(w, r)
	config := check(task, w)
	res, _ := config.Run()
	json.NewEncoder(w).Encode(res)
}
