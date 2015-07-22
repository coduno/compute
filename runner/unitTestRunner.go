package runner

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
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

// JavaUnitTest is the struct to be passed to the GeneralRun function
type JavaUnitTest struct {
}

func (s JavaUnitTest) computeResults(w http.ResponseWriter, runResults RunResults) {
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
