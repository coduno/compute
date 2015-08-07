package runner

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/coduno/compute/docker"
)

// UnitTestResult holds the unit test result created by junit
type UnitTestResult struct {
	Tests    int        `xml:"tests,attr"`
	Failures int        `xml:"failures,attr"`
	Errors   int        `xml:"errors,attr"`
	TestCase []TestCase `xml:"testcase"`
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

// JavaUnitTest is a handle function for Java unit tests. It writes the Application.java file in
//the tmp folder and  returns the docker run configuration
func JavaUnitTest(w http.ResponseWriter, r *http.Request) {
	task := decode(w, r)
	f := flag.NewFlagSet("taskFlags", flag.ContinueOnError)
	testsFlag := f.String("tests", "", "Defines the tests path")
	flags := strings.Split(task.Flags, " ")
	if len(flags) > 0 {
		if err := f.Parse(flags); err != nil {
			fmt.Printf(err.Error())
		}
	}

	if *testsFlag == "" {
		http.Error(w, "There is no test path provided", http.StatusBadRequest)
		return
	}
	c, err := docker.NewConfig(docker.NewImage(task.Language), "", task.Code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	appFilePath := path.Join(c.Volume, "src", "main", "java")
	os.MkdirAll(appFilePath, 777)

	fileName := path.Join(appFilePath, "Application.java")
	if err := ioutil.WriteFile(fileName, []byte(task.Code), 0777); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	testFilePath := path.Join(c.Volume, "src", "test", "java")
	os.MkdirAll(testFilePath, 777)

	testFile := path.Join(testFilePath, "TestApplication.java")
	testReader, err := fetchTestFile(*testsFlag)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	test, err := ioutil.ReadAll(testReader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := ioutil.WriteFile(testFile, test, 0777); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	rr, err := c.Run()

	var tests UnitTestResult
	fd, err := os.Open(path.Join(rr.Volume, testResultsFileName))
	if err != nil {
		fmt.Println(err)
		json.NewEncoder(w).Encode(rr)
		return
	}
	xml.NewDecoder(fd).Decode(&tests)

	json.NewEncoder(w).Encode(struct {
		docker.Result
		Results UnitTestResult
	}{
		Result: *rr, Results: tests,
	})
}
