package metrics

import (
	"fmt"
	"strconv"
	"time"

	"github.com/criteo/command-launcher/internal/command"
)

type extensibleMetrics struct {
	hook command.Command

	RepoName       string
	PackageName    string
	GroupName      string
	CmdName        string
	StartTimestamp time.Time
	UserPartition  uint8
}

func NewExtensibleMetricsCollector(hook command.Command) Metrics {
	return &extensibleMetrics{
		hook: hook,
	}
}

func (metrics *extensibleMetrics) Collect(uid uint8, repo string, pkg, group string, name string) error {
	if group == "" {
		return fmt.Errorf("unknown command")
	}

	metrics.RepoName = repo
	metrics.PackageName = pkg
	metrics.GroupName = group
	metrics.CmdName = name
	metrics.StartTimestamp = time.Now()
	metrics.UserPartition = uid

	return nil
}

func (metrics *extensibleMetrics) Send(cmdExitCode int, cmdError error) error {
	// call the external hook
	if metrics.hook != nil {
		errMsg := "nil"
		if cmdError != nil {
			errMsg = cmdError.Error()
		}
		duration := time.Now().UnixNano() - metrics.StartTimestamp.UnixNano()
		exitCode, _, err := metrics.hook.ExecuteWithOutput([]string{},
			metrics.RepoName,
			metrics.PackageName,
			metrics.GroupName,
			metrics.CmdName,
			strconv.Itoa(int(metrics.UserPartition)),
			strconv.Itoa(cmdExitCode),
			strconv.FormatInt(duration, 10),
			errMsg,
			strconv.FormatInt(metrics.StartTimestamp.Unix(), 10),
		)
		if err != nil || exitCode != 0 {
			return fmt.Errorf("failed to send metrics, exit code: %d, err: %v", exitCode, err)
		}
	}

	return nil
}
