// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package docker

import "io/ioutil"

// Dockerize does nothing on linux
func Dockerize(path string) (result string, err error) {
	return path, nil
}

// VolumeDir returns a temp path
func VolumeDir() (dir string, err error) {
	return ioutil.TempDir("", volumePattern)
}
