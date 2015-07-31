package runner

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/coduno/compute/docker"
)

// Handle function for an output test run. It writes the file with code in
// the tmp folder and  returns the docker run configuration.
func OutputTest(w http.ResponseWriter, r *http.Request) {
	task := decode(w, r)
	config := check(task, w)
	f := flag.NewFlagSet("taskFlags", flag.ContinueOnError)
	tests := f.String("tests", "", "Defines the tests path")

	// TODO(victorbalan): Enable the image flage when we will use it
	// image := f.String("image", "", "Defines a custom image")

	flags := strings.Split(task.Flags, " ")
	if len(flags) > 0 {
		if err := f.Parse(flags); err != nil {
			fmt.Printf(err.Error())
		}
	}

	if *tests == "" {
		http.Error(w, "There is no test path provided", http.StatusBadRequest)
		return
	}

	res, _ := config.Run()

	// FIXME(flowlo): Do this in init()
	b, err := ioutil.ReadFile(*tests)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	test := strings.Split(string(b), "\n")
	userOut := strings.Split(res.Stdout, "\n")
	var diffLines []int
	for i := 0; i < len(userOut); i++ {
		if userOut[i] != test[i] {
			diffLines = append(diffLines, i)
		}
	}

	json.NewEncoder(w).Encode(struct {
		docker.Result
		DiffLines []int
	}{
		Result: *res, DiffLines: diffLines,
	})
}
