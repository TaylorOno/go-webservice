package logging

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httputil"
)

// HttpLoggingMiddleware creates a middleware that logs the full request and response
func HttpLoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Dump request
		requestDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			slog.Error("failed to dump request", slog.String("error", err.Error()))
		} else {
			slog.Debug("HTTP Request", slog.String("dump", string(requestDump)))
		}

		// Record response
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           &bytes.Buffer{},
		}

		next(recorder, r)
		slog.Debug("HTTP Response", slog.Int("status", recorder.statusCode), slog.String("body", recorder.body.String()))
	}
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
