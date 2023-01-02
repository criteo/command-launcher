package metrics

type Metrics interface {
	Collect(uid uint8, repo string, pkg string, group string, name string) error

	Send(cmdExitCode int, cmdError error) error
}
