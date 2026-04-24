// log_init_helper.go — shared helper that initializes errlog for long-running
// commands (scan, rescan, rescan-failed). When keepLogs is false the logs
// directory is wiped so each run starts with a clean error.txt; when true the
// previous run's log is preserved (useful for diffing two consecutive runs).
package cmd

import (
	"fmt"
	"os"

	"github.com/alimtvnetwork/movie-cli-v6/errlog"
)

// initRunLogger initializes the global errlog. If keepLogs is true it appends
// to the existing log file; otherwise it wipes the logs directory first.
// commandName is only used in the warning message if init fails.
func initRunLogger(outputDir, commandName string, keepLogs bool) {
	var initErr error
	if keepLogs {
		fmt.Printf("📝 --keep-logs: appending to existing %s log\n", commandName)
		initErr = errlog.Init(outputDir, commandName)
	} else {
		initErr = errlog.InitFresh(outputDir, commandName)
	}
	if initErr != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Could not init error logger: %v\n", initErr)
	}
}
