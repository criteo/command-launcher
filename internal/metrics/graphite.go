package metrics

import (
	"fmt"
	"strconv"
	"time"

	"github.com/marpaia/graphite-golang"
)

const (
	graphitePort = 3341
)

type defaultMetrics struct {
	graphiteHost   string
	CmdName        string
	SubCmdName     string
	StartTimestamp time.Time
	UserPartition  uint8
}

func NewMetricsCollector(host string) Metrics {
	return &defaultMetrics{
		graphiteHost: host,
	}
}

func (metrics *defaultMetrics) Collect(uid uint8, cmd string, subCmd string) error {
	if cmd == "" {
		return fmt.Errorf("unknown command")
	}

	metrics.CmdName = cmd
	metrics.SubCmdName = subCmd
	metrics.StartTimestamp = time.Now()
	metrics.UserPartition = uid

	return nil
}

func (metrics *defaultMetrics) Send(cmdError error) error {
	duration := time.Now().UnixNano() - metrics.StartTimestamp.UnixNano()
	graphiteClient, err := graphite.GraphiteFactory("udp", metrics.graphiteHost, graphitePort, metrics.prefix())
	if err != nil {
		return fmt.Errorf("cannot create the graphite client: %v", err)
	}

	graphiteMetrics := []graphite.Metric{
		graphite.NewMetric("duration", strconv.FormatInt(duration, 10), metrics.StartTimestamp.Unix()),
		graphite.NewMetric("count", "1", metrics.StartTimestamp.Unix()),
	}

	if cmdError != nil {
		graphiteMetrics = append(graphiteMetrics, graphite.NewMetric("ko", "1", metrics.StartTimestamp.Unix()))
	} else {
		graphiteMetrics = append(graphiteMetrics, graphite.NewMetric("ok", "1", metrics.StartTimestamp.Unix()))
	}

	err = graphiteClient.SendMetrics(graphiteMetrics)

	return err
}

func (metrics *defaultMetrics) prefix() string {
	return fmt.Sprintf("devtools.cdt.%s.%s.%d", metrics.CmdName, metrics.SubCmdName, metrics.UserPartition)
}
