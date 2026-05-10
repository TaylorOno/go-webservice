package rest

import (
	"net/http"
	"time"
)

var (
	defaultErrorHandler = func(clientName string, method string, headers http.Header, err error) {}
	defaultStatsHandler = func(clientName string, host string, method string, headers http.Header, statusCode int, executionTime time.Duration) {
	}
)

// MetricsReporter contract for metrics provider
type MetricsReporter interface {
	RegisterCounter(name string, description string, labels ...string)
	RegisterGauge(name string, description string, labels ...string)
	RegisterSummary(name string, description string, quantiles map[float64]float64, labels ...string)
	RegisterHistogram(name string, description string, buckets []float64, labels ...string)
	IncCounter(name string, value float64, labels ...string)
	SetGauge(name string, value float64, labels ...string)
	ObserveSummary(name string, value float64, labelsValues ...string)
	ObserveHistogram(name string, value float64, labelsValues ...string)
}

// StatsHandler is a function that can be called to observer client call related metrics.
type StatsHandler func(clientName string, host string, method string, headers http.Header, statusCode int, executionTime time.Duration)

// ErrorHandler is a function that can be called to observer client call related errors.
type ErrorHandler func(clientName string, method string, headers http.Header, err error)

type ClientMiddleware func(Doer) Doer

// ClientFunc any function that can take in a *http.Request and return a *http.Response and an error can be used a ClientMiddleware.
type ClientFunc func(r *http.Request) (*http.Response, error)

// Do allow functions to implement the middleware interface.
func (f ClientFunc) Do(r *http.Request) (*http.Response, error) {
	return f(r)
}

type Client struct {
	clientName string
	OnStats    StatsHandler
	OnError    ErrorHandler
	client     Doer
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// Doer Generic http client interface
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}
