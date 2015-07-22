package docker

import (
	"io"
	"log"
	"os/exec"
)

const volumePattern string = "coduno-volume"

// DockerCmd holds the cmd to be executed as well as the executed
// image and the volume from wich it gets data
type DockerCmd struct {
	exec.Cmd
	image,
	volume string
}

// NewDockerCmd creates a command for a docker run
func NewDockerCmd(language, volume string) DockerCmd {
	image := "coduno/fingerprint-" + language
	return DockerCmd{
		*exec.Command(
			"docker",
			"run",
			"--rm",
			"-v",
			volume+":/run",
			image),
		image,
		volume,
	}
}

// GetStreams returns the in, out and err streams from a command
func GetStreams(cmd *exec.Cmd) (outCmd, errCmd io.ReadCloser, inCmd io.WriteCloser, err error) {
	outCmd, err = cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	errCmd, err = cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	inCmd, err = cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	return
}
