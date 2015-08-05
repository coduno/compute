package docker

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"syscall"
	"time"

	"google.golang.org/cloud/datastore"
)

// Config holds the configuration needed for a docker run
type Config struct {
	Image,
	Volume,
	Code,
	User string
	Challenge *datastore.Key
}

// Rusage is is a copy of syscall.Rusage for Linux.
// It is needed as syscall.Rusage built on Windows cannot be saved in Datastore
// because it holds unsigned integers. Runs are executed inside Docker and
// therefore will always generate this version of syscall.Rusage.
// See https://godoc.org/google.golang.org/cloud/datastore#Property
// See https://golang.org/src/syscall/syscall_windows.go
// See https://golang.org/src/syscall/ztypes_linux_amd64.go
type Rusage struct {
	Utime,
	Stime syscall.Timeval
	Maxrss,
	Ixrss,
	Idrss,
	Isrss,
	Minflt,
	Majflt,
	Nswap,
	Inblock,
	Oublock,
	Msgsnd,
	Msgrcv,
	Nsignals,
	Nvcsw,
	Nivcsw int64
}

const volumePattern string = "coduno-volume"
const imagePattern string = "coduno/fingerprint-"

// NewConfig constructs a Config to run the specfied image
// using the given volume.
// If volume is left blank, NewConfig will take care of
// creating a temporary directory.
func NewConfig(image, volume, code string) (c *Config, err error) {
	if volume == "" {
		volume, err = volumeDir()
	}
	c = &Config{
		Image:  image,
		Volume: volume,
		Code:   code,
	}
	return
}

// Result holds the results (standard output and standard error)
// of a Config that was run.
type Result struct {
	Config
	Stdout     string
	Stderr     string
	Rusage     Rusage
	Prepare    string
	Exit       string
	Start, End time.Time
}

// Run executes a Config and returns associated results.
func (c *Config) Run() (r *Result, err error) {
	r = &Result{Config: *c}
	dockerized, err := dockerize(c.Volume)

	if err != nil {
		return
	}

	cmd := exec.Command(
		"docker",
		"run",
		"-v",
		dockerized+":/run",
		c.Image,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return
	}
	r.Start = time.Now()
	if err = cmd.Start(); err != nil {
		return
	}

	bufout, buferr := new(bytes.Buffer), new(bytes.Buffer)

	go io.Copy(io.MultiWriter(os.Stdout, bufout), stdout)
	go io.Copy(io.MultiWriter(os.Stdout, buferr), stderr)

	err = cmd.Wait()
	if err != nil {
		r.Exit = err.Error()
	}
	r.End = time.Now()
	r.Stdout = bufout.String()
	r.Stderr = buferr.String()
	b, _ := ioutil.ReadFile(path.Join(c.Volume, "prepare.log"))
	r.Prepare = string(b)

	stats, _ := os.Open(path.Join(c.Volume, "stats.log"))
	json.NewDecoder(stats).Decode(&r.Rusage)
	return
}

// NewImage returns the correct docker image name for a
// specific language.
func NewImage(language string) string {
	return imagePattern + language
}
