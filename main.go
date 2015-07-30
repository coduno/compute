package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/coduno/compute/runner"
	"golang.org/x/net/context"
	"google.golang.org/cloud/datastore"
)

type CodeTask struct {
	Runner    string
	Flags     string
	Languages []string
}

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

func startHandler(w http.ResponseWriter, req *http.Request) {
	cors(w, req)
	encodedKey := req.URL.Path[1:]

	// TODO(victorbalan): Remove this after we can connect with the engine to localhost.
	// Untill then leave it so we can get entity keys to query for.
	// var challenges []model.Challenge
	// q := datastore.NewQuery(model.ChallengeKind).Filter("Runner =", "simple")
	// t, _ := q.GetAll(NewContext(), &challenges)
	// fmt.Println(t[0])
	// fmt.Println(t[0].Encode())
	// return

	key, err := datastore.DecodeKey(encodedKey)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var task CodeTask
	err = client.Get(context.Background(), key, &task)
	if err != nil {
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

	config := rh.Handle(w, req)
	config.Challenge = key
	config.User = "receivedUser"

	// TODO(flowlo): Find out whether Handle completed successfully
	// to check.
	res, err := config.Run(client)

	if err != nil {
		http.Error(w, "docker: "+err.Error(), http.StatusInternalServerError)
	}

	rh.Respond(w, req, res)
}

// Rudimentary CORS checking. See
// https://developer.mozilla.org/docs/Web/HTTP/Access_control_CORS
func cors(w http.ResponseWriter, req *http.Request) {
	origin := req.Header.Get("Origin")

	// only allow CORS on localhost for development
	if strings.HasPrefix(origin, "http://localhost") {
		// The cookie related headers are used for the api requests authentication
		w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST, PUT, DELETE, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Cookie, Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Origin", origin)
		if req.Method == "OPTIONS" {
			w.Write([]byte("OK"))
		}
	}
}
