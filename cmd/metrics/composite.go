package metrics

import "fmt"

type compositeMetrics struct {
	metricsList []Metrics
}

func NewCompositeMetricsCollector(list ...Metrics) Metrics {
	return &compositeMetrics{
		metricsList: list,
	}
}

func (metrics *compositeMetrics) Collect(uid uint8, repo string, pkg string, group string, name string) error {
	errPool := []error{}
	for _, m := range metrics.metricsList {
		if err := m.Collect(uid, repo, pkg, group, name); err != nil {
			errPool = append(errPool, err)
		}
	}
	if len(errPool) > 0 {
		return fmt.Errorf("multiple errors (%d), displaying first: %v", len(errPool), errPool[0])
	}
	return nil
}

func (metrics *compositeMetrics) Send(cmdExitCode int, cmdError error) error {
	errPool := []error{}
	for _, m := range metrics.metricsList {
		if err := m.Send(cmdExitCode, cmdError); err != nil {
			errPool = append(errPool, err)
		}
	}
	if len(errPool) > 0 {
		return fmt.Errorf("multiple errors (%d), displaying first: %v", len(errPool), errPool[0])
	}
	return nil

}
