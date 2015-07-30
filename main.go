package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/coduno/compute/runner"
	"golang.org/x/net/context"
	"google.golang.org/cloud/datastore"
)

const appID = "coduno"

var client *datastore.Client

func init() {
	var err error
	client, err = datastore.NewClient(context.Background(), appID)

	if err != nil {
		panic(err)
	}
}

func main() {
	http.HandleFunc("/", startHandler)
	http.ListenAndServe(":8081", nil)
}

func startHandler(w http.ResponseWriter, r *http.Request) {
	var task runner.CodeTask
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	var rh runner.RunHandler
	switch task.Runner {
	case "simple":
		rh = runner.SimpleRunHandler{}
	case "javut":
		rh = runner.JavaUnitTestHandler{}
	case "outputtest":
		if *tests == "" {
			http.Error(w, "There is no test path provided", http.StatusInternalServerError)
			return
		}
		rh = runner.OutputTestHandler{TestFilePath: *tests}
	default:
		http.Error(w, "Runner not available.", http.StatusInternalServerError)
		return
	}

	config := rh.Handle(task, w, r)
	config.User = "receivedUser"

	// TODO(flowlo): Find out whether Handle completed successfully
	// to check.
	res, err := config.Run(client)

	if err != nil {
		http.Error(w, "docker: "+err.Error(), http.StatusInternalServerError)
	}

	rh.Respond(w, r, res)
}
