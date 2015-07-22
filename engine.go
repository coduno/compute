package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/coduno/app/models"
	appUtils "github.com/coduno/app/util"
	"github.com/coduno/piper/docker"
	"github.com/coduno/piper/runner"
	"github.com/coduno/piper/util"
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

func getCodeDataFromRequest(r *http.Request) (codeData models.CodeData, err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &codeData)
	return
}

func startUnitTestRun(w http.ResponseWriter, r *http.Request) {
	if !appUtils.CheckMethod(w, r, "POST") {
		return
	}

	codeData, err := getCodeDataFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpDir, _ := runner.PrepareFilesForDockerUnitTestRun(&codeData)
	runner.UnitTestRun(w, r, tmpDir, &codeData)
}

func startSimpleRun(w http.ResponseWriter, r *http.Request) {
	if !appUtils.CheckMethod(w, r, "POST") {
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
	err = util.CreateFile(tmpDir, fileNames[codeData.Language], codeData.CodeBase)
	if err != nil {
		http.Error(w, "File preparation error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	runner.SimpleRun(w, r, tmpDir, &codeData)
}

func main() {
	http.HandleFunc("/api/run/start/simple", startSimpleRun)
	http.HandleFunc("/api/run/start/unittest", startUnitTestRun)
	http.ListenAndServe(":8081", nil)
}
