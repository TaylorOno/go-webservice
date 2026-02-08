package metrics

import (
	"net/http"
	"strconv"
	"time"
)

const (
	_incomingReqHist    = "app_request_latency_histogram"
	_incomingReqSummary = "app_request_latency"
)

type Registry interface {
	RegisterHistogram(name string, description string, buckets []float64, labels ...string)
	RegisterSummary(name string, description string, quantiles map[float64]float64, labels ...string)
	ObserveHistogram(name string, value float64, labels ...string)
	ObserveSummary(name string, value float64, labels ...string)
}

func HttpMiddleware(registry Registry) func(next http.HandlerFunc) http.HandlerFunc {
	registry.RegisterHistogram(_incomingReqHist, "Service response time", []float64{1, 3, 5, 10, 25, 50, 100, 200, 400, 600, 1000, 1500, 2000, 2500, 3000, 4000, 5000, 10000}, "path", "method")
	registry.RegisterSummary(_incomingReqSummary, "Service response time with more labels", map[float64]float64{}, "path", "method", "status_code")

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			responseRecorder := newResponseRecorder(w)

			path := r.URL.Path
			defer func(start time.Time) {
				registry.ObserveHistogram(_incomingReqHist, ToMilliseconds(time.Since(start)), r.Method, path, strconv.Itoa(responseRecorder.statusCode))
				registry.ObserveSummary(_incomingReqSummary, ToMilliseconds(time.Since(start)), r.Method, path, strconv.Itoa(responseRecorder.statusCode))
			}(time.Now())

			next.ServeHTTP(responseRecorder, r)
		}
	}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{w, http.StatusOK}
}

func (lrw *responseRecorder) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
