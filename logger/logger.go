package logger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"syscall"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/cloud/datastore"
)

// Rusage is the copy from syscall.Rusage for Linux.
// It is needed because without it the LogRunInfo would fail.
// The reason is that we would try to unmarshal an linux Rusage into
// an windows Rusage
type Rusage struct {
	Utime    syscall.Timeval
	Stime    syscall.Timeval
	Maxrss   int64
	Ixrss    int64
	Idrss    int64
	Isrss    int64
	Minflt   int64
	Majflt   int64
	Nswap    int64
	Inblock  int64
	Oublock  int64
	Msgsnd   int64
	Msgrcv   int64
	Nsignals int64
	Nvcsw    int64
	Nivcsw   int64
}

// RunLog is used to represent general data for a single invocation of
// a Coduno build
type RunLog struct {
	Challenge *datastore.Key
	User      string
	Code      string
	Status    string
	StartTime time.Time
	EndTime   time.Time
}

// RunInfo is used to represent accumulated log data of a single invocation of
// a Coduno build
type RunInfo struct {
	RunLog     *datastore.Key
	InLog      string `datastore:",noindex"`
	OutLog     string `datastore:",noindex"`
	ErrLog     string `datastore:",noindex"`
	PrepareLog string `datastore:",noindex"`
	SysUsage   Rusage
}

const runLogKind = "runLog"
const runInfoKind = "runInfo"

// LogRunStart sends info to the datastore, informing that a new run
// started
func LogRunStart(client *datastore.Client, challenge *datastore.Key, code string, user string) (*datastore.Key, *RunLog) {
	ctx := context.Background()
	key := datastore.NewIncompleteKey(ctx, runLogKind, nil)
	log := &RunLog{challenge, user, code, "started", time.Now(), time.Unix(0, 0)}

	key, err := client.Put(ctx, key, log)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LogBuildStart: %v", err)
	}
	return key, log
}

// LogRunComplete logs the end of a completed (failed of finished) run of
// a coduno testrun
func LogRunComplete(client *datastore.Client, runLogKey *datastore.Key, log *RunLog, exit error) {
	ctx := context.Background()
	log.EndTime = time.Now()
	if exit != nil {
		log.Status = "failed"
	} else {
		log.Status = "good"
	}
	_, err := client.Put(ctx, runLogKey, log)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}

// LogRunInfo logs info of a completed run
func LogRunInfo(client *datastore.Client, runLogKey *datastore.Key, in,
	out, cmdErr, prepLog string, stats Rusage) {
	ctx := context.Background()

	data := &RunInfo{runLogKey, in, out, cmdErr, prepLog, stats}
	key := datastore.NewIncompleteKey(ctx, runInfoKind, nil)
	_, err := client.Put(ctx, key, data)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

}

// GetLogs gets the prepare and the stats logs
func GetLogs(tmpDir string) (string, Rusage) {
	prepLog, err := ioutil.ReadFile(tmpDir + "/prepare.log")
	if err != nil {
		log.Print(err)
	}

	var stats Rusage
	statsData, err := ioutil.ReadFile(tmpDir + "/stats.log")
	if err != nil {
		log.Print(err)
	} else {
		err = json.Unmarshal(statsData, &stats)
		if err != nil {
			log.Print(err)
		}
	}

	return string(prepLog), stats
}
