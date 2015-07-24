package runner

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"path"

	"github.com/coduno/piper/docker"
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

// JavaUnitTestHandler is the handler for java unit tests
type JavaUnitTestHandler struct {
	tmpDir string
}

// Handle function for Java unit tests. It writes the Application.java file in
//the tmp folder and  returns the docker run configuration
func (jut *JavaUnitTestHandler) Handle(w http.ResponseWriter, r *http.Request) (c docker.Config) {
	// TODO(victorbalan): POST Method check

	codeData, err := getCodeDataFromRequest(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c, err = docker.NewConfig(docker.NewImage(codeData.Language))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = ioutil.WriteFile(path.Join(c.Volume, "Application.java"), []byte(codeData.CodeBase), 0777)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return
}

// Respond implementation for Java unit tests. Send the JUnit results too.
func (jut *JavaUnitTestHandler) Respond(w http.ResponseWriter, r *http.Request, rr docker.Result) {
	testResult, err := ioutil.ReadFile(path.Join(jut.tmpDir, testResultsFileName))
	if err != nil {
		log.Print(err)
	}

	var result UnitTestResult
	err = xml.Unmarshal(testResult, &result)
	var toSend = make(map[string]interface{})
	toSend["run"] = rr.Stdout
	toSend["err"] = rr.Stderr
	toSend["tests"] = result
	json, err := json.Marshal(toSend)
	if err != nil {
		http.Error(w, "Json marshal err: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(json)
}
