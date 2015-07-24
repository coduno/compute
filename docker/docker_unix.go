// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package docker

import "io/ioutil"

// dockerize does nothing on linux
func dockerize(path string) (result string, err error) {
	return path, nil
}

// volumeDir returns a temp path
func volumeDir() (dir string, err error) {
	return ioutil.TempDir("", volumePattern)
}
