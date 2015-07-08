package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path"
	"sync"
	"syscall"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/cloud"
	"google.golang.org/cloud/datastore"
)

// BuildData is used to represent general data for a single invocation of
// a Coduno build
type BuildData struct {
	Challenge string
	User      string
	Commit    string
	Status    string
	StartTime time.Time
	EndTime   time.Time
}

// LogData is used to represent accumulated log data of a single invocation of
// a Coduno build
type LogData struct {
	InLog      string `datastore:",noindex"`
	OutLog     string `datastore:",noindex"`
	ExtraLog   string `datastore:",noindex"`
	PrepareLog string `datastore:",noindex"`
	SysUsage   syscall.Rusage
}

// PipeStatus holds the satus of a buffered pipe.
// It tells how many bytes have been read from the source,
// wrote to the destination and were buffered.
type PipeStatus struct {
	Read,
	Wrote,
	Buffered int

	ReadError,
	WriteError,
	BufferError error
}

const buildKind = "builds"

var ctx context.Context

func init() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	fileName := path.Join(user.HomeDir, "config", "secret.json")
	secret, err := ioutil.ReadFile(fileName)
	if err != nil {
		panic(err)
	}

	config, err := google.JWTConfigFromJSON(secret, datastore.ScopeDatastore, datastore.ScopeUserEmail)
	if err != nil {
		panic(err)
	}

	ctx = cloud.NewContext("coduno", config.Client(oauth2.NoContext))
}

// LogBuildStart sends info to the datastore, informing that a new build
// started
func LogBuildStart(repo string, commit string, user string) (*datastore.Key, *BuildData) {
	key := datastore.NewIncompleteKey(ctx, buildKind, nil)
	build := &BuildData{repo, user, commit, "started", time.Now(), time.Unix(0, 0)}

	key, err := datastore.Put(ctx, key, build)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LogBuildStart: %v", err)
	}
	return key, build
}

// LogRunComplete logs the end of a completed (failed of finished) run of
// a coduno testrun
func LogRunComplete(pKey *datastore.Key, build *BuildData, in,
	out, extra string, exit error, prepLog string, stats syscall.Rusage) {
	tx, err := datastore.NewTransaction(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LogRunComplete: Could not get transaction!")
	}
	build.EndTime = time.Now()
	if exit != nil {
		build.Status = "failed"
	} else {
		build.Status = "good"
	}
	_, err = tx.Put(pKey, build)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LogRunComplete: Putting build failed!")
		tx.Rollback()
		return
	}
	data := &LogData{in, out, extra, prepLog, stats}
	k := datastore.NewIncompleteKey(ctx, buildKind, pKey)
	_, err = tx.Put(k, data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LogRunComplete: Putting data failed!")
		tx.Rollback()
		return
	}
	tx.Commit()
}

func pipeOutput(wg *sync.WaitGroup, rc io.ReadCloser, w io.Writer, buf *bytes.Buffer) (parent chan<- PipeStatus) {
	defer wg.Done()
	// if we have no rc we cannot do anything, because
	// that's where the data come from
	if rc == nil {
		return
	}
	defer rc.Close()

	s := PipeStatus{}

	tmp := make([]byte, 1024)

	// to count how many bytes we read/write/buffer on
	// every loop
	var cR, cW, cB int

	for s.ReadError == nil && (s.WriteError == nil || s.BufferError == nil) {
		cR, s.ReadError = rc.Read(tmp)
		s.Read += cR

		if cR == 0 {
			continue
		}

		if buf != nil && s.BufferError == nil {
			cB, s.BufferError = buf.Write(tmp[0:cR])
			s.Buffered += cB
		}

		if w != nil && s.WriteError == nil {
			cW, s.WriteError = w.Write(tmp[0:cR])
			s.Wrote += cW
		}
	}
	parent <- s
	return
}

func main() {
	if len(os.Args) != 6 {
		log.Fatal(os.Args)
	}

	username := os.Args[1]
	repo := os.Args[2]
	commit := os.Args[3]
	tmpdir := os.Args[4]
	testdir := os.Args[5]

	key, build := LogBuildStart(repo, commit, username)
	cmdUser := exec.Command(
		"sudo",
		"docker",
		"run",
		"--rm",
		"-v",
		tmpdir+":/run",
		"coduno/base")
	cmdTest := exec.Command(
		"sudo",
		"docker",
		"run",
		"--rm",
		"-v",
		testdir+":/run",
		"coduno/base")

	outUser, err := cmdUser.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	errUser, err := cmdUser.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	inUser, err := cmdUser.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	outTest, err := cmdTest.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	errTest, err := cmdTest.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	inTest, err := cmdTest.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	var userToTest, testToUser, extraBuf bytes.Buffer
	cmdUser.Start()
	cmdTest.Start()

	var wg sync.WaitGroup
	wg.Add(4)

	go pipeOutput(&wg, outUser, inTest, &userToTest)
	go pipeOutput(&wg, outTest, inUser, &testToUser)
	go pipeOutput(&wg, errUser, os.Stderr, nil)
	go pipeOutput(&wg, errTest, nil, &extraBuf)

	exitErr := cmdUser.Wait()
	wg.Wait()

	prepLog, err := ioutil.ReadFile(tmpdir + "/prepare.log")
	if err != nil {
		log.Fatal(err)
	}

	var stats syscall.Rusage
	statsData, err := ioutil.ReadFile(tmpdir + "/stats.log")
	if err != nil {
		log.Print(err)
	} else {
		err = json.Unmarshal(statsData, &stats)
		if err != nil {
			log.Fatal(err)
		}
	}

	LogRunComplete(key, build, testToUser.String(), userToTest.String(), extraBuf.String(), exitErr, string(prepLog), stats)
}
