package runner

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/storage/v1"

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

const codunoComputeStorage = "CODUNO_COMPUTE_STORAGE"

func fetchTestFile(objectName string) (reader io.ReadCloser, err error) {
	localStorage := os.Getenv(codunoComputeStorage)
	if localStorage == "" {
		bucketName := "task_tests"
		var client *http.Client
		if client, err = google.DefaultClient(context.Background(), storage.DevstorageReadOnlyScope); err != nil {
			return
		}

		var service *storage.Service
		if service, err = storage.New(client); err != nil {
			return
		}

		var res *storage.Object
		if res, err = service.Objects.Get(bucketName, objectName).Do(); err != nil {
			return
		}

		var fileDld *http.Response
		if fileDld, err = client.Get(res.MediaLink); err != nil {
			return
		}
		// TODO(victorbalan): Put in memcache
		io.Copy(os.Stdout, fileDld.Body)
		return fileDld.Body, nil
	}
	if reader, err = os.Open(path.Join(localStorage, objectName)); err != nil {
		return
	}
	return
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
