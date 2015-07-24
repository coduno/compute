package docker

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

const volumePattern string = "coduno-volume"
const imagePattern string = "coduno/fingerprint-"

// Config holds the configuration needed for a docker run
type Config struct {
	Image  string
	Volume string
}

// NewConfig constructs a Config to run the specfied image,
// automatically taking care of allocating a directory for
// Volume.
func NewConfig(image string) (c Config, err error) {
	volume, err := volumeDir()
	c = Config{
		image,
		volume,
	}
	return
}

// Result holds the results (standard output and standard error)
// of a Config that was run.
type Result struct {
	Config
	Stdout bytes.Buffer
	Stderr bytes.Buffer
}

// Run executes a Config and returns associated results.
func (c Config) Run() (r Result, err error) {
	dockerized, err := dockerize(c.Volume)

	if err != nil {
		return
	}

	cmd := exec.Command(
		"docker",
		"run",
		"--rm",
		"-v",
		dockerized+":/run",
		c.Image)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return
	}

	if err = cmd.Start(); err != nil {
		return
	}

	go io.Copy(io.MultiWriter(os.Stdout, &r.Stdout), stdout)
	go io.Copy(io.MultiWriter(os.Stdout, &r.Stderr), stderr)

	cmd.Wait()
	return
}

// NewImage returns the correct docker image name for a
// specific language.
func NewImage(language string) string {
	return imagePattern + language
}
