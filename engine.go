package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/coduno/app/models"
	"github.com/coduno/app/util"
	"github.com/coduno/piper/docker"
	"github.com/coduno/piper/runner"
)

var (
	fileNames = map[string]string{
		"py":   "app.py",
		"c":    "app.c",
		"cpp":  "app.cpp",
		"java": "Application.java",
	}
)

const configFileName string = "coduno.yaml"

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

func getCodeDataFromRequest(r *http.Request) (codeData models.CodeData, err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &codeData)
	return
}

func startUnitTestRun(w http.ResponseWriter, r *http.Request) {
	if !util.CheckMethod(w, r, "POST") {
		return
	}

	codeData, err := getCodeDataFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	switch codeData.Language {
	case "javaut":
		tmpDir, err := docker.VolumeDir()
		if err != nil {
			http.Error(w, "Volume preparation error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		err = ioutil.WriteFile(path.Join(tmpDir, "Application.java"), []byte(codeData.CodeBase), 0777)
		if err != nil {
			http.Error(w, "File preparation error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		runner.GeneralRun(w, r, tmpDir, &codeData, runner.JavaUnitTest{})
	default:
		http.Error(w, "Language not available for unit testing", http.StatusBadRequest)
	}
}

func startSimpleRun(w http.ResponseWriter, r *http.Request) {
	cors(w, r)
	if !util.CheckMethod(w, r, "POST") {
		return
	}

	codeData, err := getCodeDataFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	for availableLanguage := range fileNames {
		if codeData.Language == availableLanguage {
			goto LANGUAGE_AVAILABLE
		}
	}
	http.Error(w, "Language not available.", http.StatusBadRequest)

LANGUAGE_AVAILABLE:
	tmpDir, err := docker.VolumeDir()
	if err != nil {
		http.Error(w, "Volume preparation error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	err = ioutil.WriteFile(path.Join(tmpDir, fileNames[codeData.Language]), []byte(codeData.CodeBase), 0777)
	if err != nil {
		http.Error(w, "File preparation error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	runner.GeneralRun(w, r, tmpDir, &codeData, runner.SimpleRun{})
}

func main() {
	http.HandleFunc("/api/run/start/simple", startSimpleRun)
	http.HandleFunc("/api/run/start/unittest", startUnitTestRun)
	http.ListenAndServe(":8081", nil)
}
