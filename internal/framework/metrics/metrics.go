package metrics

import "strings"

const (
	metricTypeCounter   = "counter"
	metricTypeGauge     = "gauge"
	metricTypeHistogram = "histogram"
	metricTypeSummary   = "summary"
)

var (
	sanitizer      = strings.NewReplacer(".", "_", "-", "_")
	defaultBuckets = []float64{0, 5, 10, 25, 50, 75, 100, 250, 500, 750, 1000, 2500, 5000, 7500, 10000}
)

type MetricDefinition struct {
	Kind        string
	Description string
	labelCount  int
	Labels      []string
}

func sanitize(labels []string) {
	for i, label := range labels {
		sanitizedLabels := sanitizer.Replace(label)
		labels[i] = sanitizedLabels
	}
}
