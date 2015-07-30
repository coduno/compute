package runner

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/coduno/compute/docker"
	"github.com/coduno/engine/util"
)

// OutputTestHandler is the handler for a run with a simple output check
type OutputTestHandler struct {
	TestFilePath string
}

// Handle function for an output test run. It writes the file with code in
// the tmp folder and  returns the docker run configuration.
func (oth OutputTestHandler) Handle(task CodeTask, w http.ResponseWriter, r *http.Request) (c docker.Config) {
	return GeneralHandle(task, w, r)
}

// Respond implementation for an output test run. It returns the number of
// different lines between the user output, the user output and the test
// output.
func (oth OutputTestHandler) Respond(w http.ResponseWriter, req *http.Request, res docker.Result) {
	buf, err := ioutil.ReadFile(oth.TestFilePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	test := strings.Split(string(buf), "\n")
	cmdErr := res.Stderr.String()

	if cmdErr != "" {
		util.WriteMap(w, map[string]interface{}{
			"stderr": cmdErr,
		})
		return
	}
	userOut := strings.Split(res.Stdout.String(), "\n")
	if len(test) != len(userOut) {
		util.WriteMap(w, map[string]interface{}{
			"stdout": res.Stdout.String(),
		})
		return
	}

	var diffLines []int
	for i := 0; i < len(userOut); i++ {
		if userOut[i] != test[i] {
			diffLines = append(diffLines, i)
		}
	}
	util.WriteMap(w, map[string]interface{}{
		"stdout":    string(res.Stdout.Bytes()),
		"diffLines": diffLines,
	})
}
