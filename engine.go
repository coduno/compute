package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/coduno/piper/runner"
)

var code = flag.String("code", "", "location of submitted code")

func setupRunHandler(rh runner.RunHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cors(w, r)
		config := rh.Handle(w, r)

		// TODO(flowlo): Find out whether Handle completed successfully
		// to check.

		res, err := config.Run()

		if err != nil {
			http.Error(w, "docker: "+err.Error(), http.StatusInternalServerError)
		}

		rh.Respond(w, r, res)
	}
}

func main() {
	flag.Parse()

	if *code != "" {
		srh := runner.SimpleRunHandler{}
		config := srh.Prepare(*code)

		_, err := config.Run()

		if err != nil {
			fmt.Fprintf(os.Stderr, "docker: "+err.Error())
		}
	} else {
		serve()
	}
}

func serve() {
	http.HandleFunc("/api/run/start/simple", setupRunHandler(runner.SimpleRunHandler{}))
	http.HandleFunc("/api/run/start/unittest", setupRunHandler(&runner.JavaUnitTestHandler{}))
	http.ListenAndServe(":8081", nil)
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
