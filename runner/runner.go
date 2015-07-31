package runner

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"path"

	"github.com/coduno/compute/docker"
)

var (
	fileNames = map[string]string{
		"py":   "app.py",
		"c":    "app.c",
		"cpp":  "app.cpp",
		"java": "Application.java",
	}
)

type CodeTask struct {
	Flags, Code, Runner, Language string
}

func decode(w http.ResponseWriter, r *http.Request) *CodeTask {
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return nil
	}

	var task CodeTask
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return nil
	}
	return &task
}

func check(task *CodeTask, w http.ResponseWriter) (c *docker.Config) {
	for availableLanguage := range fileNames {
		if task.Language == availableLanguage {
			goto LANGUAGE_AVAILABLE
		}
	}
	http.Error(w, "language not available", http.StatusBadRequest)
	return

LANGUAGE_AVAILABLE:
	c, err := docker.NewConfig(docker.NewImage(task.Language), "", task.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileName := path.Join(c.Volume, fileNames[task.Language])
	if err := ioutil.WriteFile(fileName, []byte(task.Code), 0777); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return
}
