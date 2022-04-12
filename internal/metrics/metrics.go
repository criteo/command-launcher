package metrics

type Metrics interface {
	Collect(uid uint8, cmd string, subCmd string) error

	Send(cmdError error) error
}
