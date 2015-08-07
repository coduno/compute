package runner

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"golang.org/x/net/context"
	"google.golang.org/cloud/storage"

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
	Flags, Code, Language string
}

const codunoComputeStorage = "CODUNO_COMPUTE_STORAGE"
const bucket = "task_tests"

var cache string

func init() {
	var err error

	if cache = os.Getenv(codunoComputeStorage); cache != "" {
		// TODO(flowlo): Maybe check if we can even read/write from/to that directory.
		return
	}

	if cache, err = ioutil.TempDir("", "coduno-compute-cache-"); err != nil {
		panic(err)
	}
}

// OpenTestFile is like os.Open but for accessing test files.
func OpenTestFile(name string) (io.Reader, error) {
	fn := path.Join(cache, name)

	f, err := os.Open(fn)
	if err == nil {
		return f, nil
	}

	if err != os.ErrNotExist {
		return nil, err
	}

	rc, err := storage.NewReader(context.Background(), bucket, name)
	if err != nil {
		return nil, err
	}

	w, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY|os.O_EXCL, os.ModeTemporary)
	if err != nil {
		return nil, err
	}

	return io.TeeReader(rc, w), nil
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
