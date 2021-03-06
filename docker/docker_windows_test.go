package docker

import "testing"

func TestDockerize(t *testing.T) {
	const expected = "/c/users/Foo"

	var dockerized string
	dockerized, err := Dockerize(`C:\users\Foo`)

	if err != nil {
		t.Error("Unexpected error: ", err)
	}

	if dockerized != expected {
		t.Error("Expected '", expected, "', got '", dockerized, "'")
	}
}
