package docker

import (
	"io"
	"os/exec"
)

const volumePattern string = "coduno-volume"
const imagePattern string = "coduno/fingerprint-"

// DockerConfig holds the configuration needed for a docker run
type DockerConfig struct {
	Image  string
	TmpDir string
	Volume string
}

// NewDockerCmd creates a command for a docker run
func NewDockerCmd(image, volume string) *exec.Cmd {
	return exec.Command(
		"docker",
		"run",
		"--rm",
		"-v",
		volume+":/run",
		image)

}

// GetImageForLanguage returns the correct docker image name for a specific
// language
func GetImageForLanguage(language string) string {
	return imagePattern + language
}

// GetStreams returns the in, out and err streams from a command
func GetStreams(cmd *exec.Cmd) (outCmd, errCmd io.ReadCloser, inCmd io.WriteCloser, err error) {
	outCmd, err = cmd.StdoutPipe()
	if err != nil {
		return
	}
	errCmd, err = cmd.StderrPipe()
	if err != nil {
		return
	}
	inCmd, err = cmd.StdinPipe()
	if err != nil {
		return
	}

	return
}
