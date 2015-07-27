package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os/user"
	"path"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/cloud"
	"google.golang.org/cloud/datastore"

	"github.com/coduno/piper/models"
	"github.com/coduno/piper/runner"
)

var ctx context.Context

func init() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	fileName := path.Join(user.HomeDir, "config", "secret.json")
	secret, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	config, err := google.JWTConfigFromJSON(secret, datastore.ScopeDatastore, datastore.ScopeUserEmail)
	if err != nil {
		panic(err)
	}

	ctx = cloud.NewContext("coduno", config.Client(oauth2.NoContext))
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
	// q := datastore.NewQuery(models.ChallengeKind).Filter("Runner =", "outputtest")
	// var challenges []models.Challenge
	// t, _ := q.GetAll(ctx, &challenges)
	// fmt.Println(t[0])
	// fmt.Println(t[0].Encode())
	// return
	// cKey = t[0].Encode()

	key, err := datastore.DecodeKey(encodedKey)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var challenge models.Challenge
	err = datastore.Get(ctx, key, &challenge)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	f := flag.NewFlagSet("challengeFlags", flag.ContinueOnError)
	tests := f.String("tests", "", "Defines the tests path")
	// TODO(victorbalan): Enable the image flage when we will use it
	// image := f.String("image", "", "Defines a custom image")

	flags := strings.Split(challenge.Flags, " ")
	if len(flags) > 0 {
		if err := f.Parse(flags); err != nil {
			fmt.Printf(err.Error())
		}
	}

	var rh runner.RunHandler
	switch challenge.Runner {
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

	// TODO(flowlo): Find out whether Handle completed successfully
	// to check.
	res, err := config.Run()

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
		w.Header().Set("Access-Control-Allow-Methods", "OPTIONS,GET,POST,PUT,DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "cookie,content-type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Origin", origin)
		if req.Method == "OPTIONS" {
			w.WriteHeader(200)
			w.Write([]byte("OK"))
		}
	}
}
