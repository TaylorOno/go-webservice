package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusReporter struct {
	sync.RWMutex
	metricRegistry       map[string]MetricDefinition
	counterDefinitions   map[string]*prometheus.CounterVec
	gaugeDefinitions     map[string]*prometheus.GaugeVec
	summaryDefinitions   map[string]*prometheus.SummaryVec
	histogramDefinitions map[string]*prometheus.HistogramVec
}

func NewPrometheusReporter() *PrometheusReporter {
	return &PrometheusReporter{
		metricRegistry:       make(map[string]MetricDefinition),
		counterDefinitions:   make(map[string]*prometheus.CounterVec),
		gaugeDefinitions:     make(map[string]*prometheus.GaugeVec),
		summaryDefinitions:   make(map[string]*prometheus.SummaryVec),
		histogramDefinitions: make(map[string]*prometheus.HistogramVec),
	}
}

func (r *PrometheusReporter) registerMetrics(name string, description string, kind string, labels []string) {
	r.Lock()
	r.metricRegistry[name] = MetricDefinition{
		Kind:        kind,
		Description: description,
		Labels:      labels,
		labelCount:  len(labels),
	}
	r.Unlock()
}

func (r *PrometheusReporter) RegisterCounter(name string, description string, labels ...string) {
	sanitize(labels)

	opts := prometheus.CounterOpts{Name: name, Help: description}
	counter := prometheus.NewCounterVec(opts, labels)
	r.registerMetrics(name, description, metricTypeCounter, labels)

	r.Lock()
	prometheus.MustRegister(counter)
	r.counterDefinitions[name] = counter
	r.Unlock()
}

func (r *PrometheusReporter) RegisterGauge(name string, description string, labels ...string) {
	sanitize(labels)

	opts := prometheus.GaugeOpts{Name: name, Help: description}
	gauge := prometheus.NewGaugeVec(opts, labels)
	r.registerMetrics(name, description, metricTypeGauge, labels)

	r.Lock()
	prometheus.MustRegister(gauge)
	r.gaugeDefinitions[name] = gauge
	r.Unlock()
}

func (r *PrometheusReporter) RegisterSummary(name string, description string, quantiles map[float64]float64, labels ...string) {
	sanitize(labels)

	opts := prometheus.SummaryOpts{Name: name, Help: description, Objectives: quantiles}
	summary := prometheus.NewSummaryVec(opts, labels)
	r.registerMetrics(name, description, metricTypeSummary, labels)

	r.Lock()
	prometheus.MustRegister(summary)
	r.summaryDefinitions[name] = summary
	r.Unlock()
}

func (r *PrometheusReporter) RegisterHistogram(name string, description string, buckets []float64, labels ...string) {
	sanitize(labels)

	if len(buckets) == 0 {
		buckets = prometheus.DefBuckets
	}

	opts := prometheus.HistogramOpts{Name: name, Help: description, Buckets: buckets}
	histogram := prometheus.NewHistogramVec(opts, labels)
	r.registerMetrics(name, description, metricTypeHistogram, labels)

	r.Lock()
	prometheus.MustRegister(histogram)
	r.histogramDefinitions[name] = histogram
	r.Unlock()
}

func (r *PrometheusReporter) IncCounter(name string, value float64, labels ...string) {
	r.RLock()
	if metric, ok := r.counterDefinitions[name]; ok {
		labelCount := r.metricRegistry[name].labelCount
		if labelCount == len(labels) {
			metric.WithLabelValues(labels...).Add(value)
		} else {
			// Error
		}
	}
	r.RUnlock()
}

func (r *PrometheusReporter) SetGauge(name string, value float64, labels ...string) {
	r.RLock()
	if metric, ok := r.gaugeDefinitions[name]; ok {
		labelCount := r.metricRegistry[name].labelCount
		if labelCount == len(labels) {
			metric.WithLabelValues(labels...).Set(value)
		}
	} else {
		// Error
	}
	r.RUnlock()
}

func (r *PrometheusReporter) ObserveSummary(name string, value float64, labels ...string) {
	r.RLock()
	if metric, ok := r.summaryDefinitions[name]; ok {
		labelCount := r.metricRegistry[name].labelCount
		if labelCount == len(labels) {
			metric.WithLabelValues(labels...).Observe(value)
		}
	} else {
		// Error
	}
	r.RUnlock()
}

func (r *PrometheusReporter) ObserveHistogram(name string, value float64, labels ...string) {
	r.RLock()
	if metric, ok := r.histogramDefinitions[name]; ok {
		labelCount := r.metricRegistry[name].labelCount
		if labelCount == len(labels) {
			metric.WithLabelValues(labels...).Observe(value)
		}
	} else {
		// Error
	}
	r.RUnlock()
}
