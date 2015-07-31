package runner

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/coduno/compute/docker"
)

// UnitTestResult holds the unit test result created by junit
type UnitTestResult struct {
	Tests    string   `xml:"tests,attr"`
	Failures string   `xml:"failures,attr"`
	Errors   string   `xml:"errors,attr"`
	TestCase TestCase `xml:"testcase"`
}

// TestCase holds a test case created by junit
type TestCase struct {
	Name    string  `xml:"name,attr"`
	Time    string  `xml:"time,attr"`
	Failure Failure `xml:"failure"`
}

// Failure holds a failure created by junit
type Failure struct {
	Message string `xml:"message,attr"`
}

const testResultsFileName = "/build/test-results/TEST-com.coduno.TestApplication.xml"

// Handle function for Java unit tests. It writes the Application.java file in
//the tmp folder and  returns the docker run configuration
func JavaUnitTest(w http.ResponseWriter, r *http.Request) {
	task := decode(w, r)

	c, err := docker.NewConfig(docker.NewImage(task.Language), "", task.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileName := path.Join(c.Volume, "Application.java")
	if err := ioutil.WriteFile(fileName, []byte(task.Code), 0777); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	rr, err := c.Run()

	var tests UnitTestResult
	fd, err := os.Open(path.Join(rr.Volume, testResultsFileName))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	xml.NewDecoder(fd).Decode(&tests)

	json.NewEncoder(w).Encode(struct {
		docker.Result
		Tests UnitTestResult
	}{
		Result: *rr, Tests: tests,
	})
}
