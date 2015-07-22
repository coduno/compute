package logger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path"
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
func LogBuildStart(challenge string, commit string, user string) (*datastore.Key, *BuildData) {
	key := datastore.NewIncompleteKey(ctx, buildKind, nil)
	build := &BuildData{challenge, user, commit, "started", time.Now(), time.Unix(0, 0)}

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

// GetLogs gets the prepare and the stats logs
func GetLogs(tmpDir string) (string, syscall.Rusage) {
	prepLog, err := ioutil.ReadFile(tmpDir + "/prepare.log")
	if err != nil {
		log.Print(err)
	}

	var stats syscall.Rusage
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
