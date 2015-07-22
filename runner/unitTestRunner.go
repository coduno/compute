package runner

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/coduno/app/models"
	"github.com/coduno/piper/docker"
	"github.com/coduno/piper/util"
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

// RunUnitTest is the struct to be passed to the GeneralRun function
type RunUnitTest struct {
}

func (s RunUnitTest) computeResults(w http.ResponseWriter, runResults RunResults) {
	testResult, err := ioutil.ReadFile(runResults.tmpDir + testResultsFileName)
	if err != nil {
		log.Print(err)
	}

	var result UnitTestResult
	err = xml.Unmarshal(testResult, &result)
	var toSend = make(map[string]interface{})
	toSend["run"] = runResults.runOut
	toSend["err"] = runResults.runErr
	toSend["tests"] = result
	json, err := json.Marshal(toSend)
	if err != nil {
		http.Error(w, "Json marshal err: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(json)
}

// UnitTestRun describes a run that sends to the client the unit test results
func UnitTestRun(w http.ResponseWriter, r *http.Request, tmpDir string, codeData *models.CodeData) {
	var run RunUnitTest
	GeneralRun(w, r, tmpDir, codeData, run)
}

// PrepareFilesForDockerUnitTestRun creates the necessary files for a unit test run
// - build.gradle
// - src/test/java/TEST_FILE
// - src/main/java/TO_TEST_FILE
func PrepareFilesForDockerUnitTestRun(codeData *models.CodeData) (tempDir string, err error) {
	tempDir, err = docker.VolumeDir()
	if err != nil {
		return
	}
	// TODO(victorbalan): use the real unit tests
	appPath := path.Join(tempDir, "src", "main", "java")
	os.MkdirAll(appPath, 0777)
	err = util.CreateFile(appPath, "Application.java", codeData.CodeBase)
	if err != nil {
		return
	}

	err = util.CopyFileContents(tempDir, "unit_test/build.gradle", "build.gradle")
	if err != nil {
		return
	}

	testPath := path.Join(tempDir, "src", "test", "java")
	err = util.CopyFileContents(testPath, "unit_test/TestApplication.java", "TestApplication.java")
	if err != nil {
		return
	}

	return tempDir, nil
}
