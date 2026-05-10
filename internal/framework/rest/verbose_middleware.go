package rest

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"strings"
)

type VerboseMiddleware struct {
	level slog.Level
}

func Verbose(level slog.Level) *VerboseMiddleware {
	return &VerboseMiddleware{level: level}
}

// RequestLogger prints basic request information to standard output
func (v *VerboseMiddleware) RequestLogger() ClientMiddleware {
	return func(c Doer) Doer {
		return ClientFunc(func(r *http.Request) (*http.Response, error) {
			v.logRequest(r)
			res, err := c.Do(r)
			v.logResponse(res)
			return res, err
		})
	}
}

func (v *VerboseMiddleware) logRequest(r *http.Request) {
	if v.level < -4 {
		v.traceRequest(r)
	}
}

func (v *VerboseMiddleware) traceRequest(r *http.Request) {
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		slog.Error("failed to dump request", slog.String("error", err.Error()))
		return
	}

	// We need to parse the dump to separate headers and body
	parts := strings.SplitN(string(requestDump), "\r\n\r\n", 2)
	if len(parts) != 2 {
		slog.Log(r.Context(), v.level, "HTTP Request", slog.String("dump", string(requestDump)))
		return
	}

	if indented, ok := prettyJSON([]byte(parts[1])); ok {
		slog.Log(r.Context(), v.level, "HTTP Request", slog.String("headers", parts[0]), slog.String("body", indented))
		return
	}

	slog.Log(r.Context(), v.level, "HTTP Request", slog.String("dump", string(requestDump)))
}

func (v *VerboseMiddleware) logResponse(recorder *http.Response) {
	switch {
	case v.level < -4:
		v.traceResponse(recorder)
	case v.level < -0:
		v.debugResponse(recorder)
	}
}

func (v *VerboseMiddleware) debugResponse(r *http.Response) {
	slog.Log(r.Request.Context(), v.level, "HTTP Response",
		slog.String("method", r.Request.Method),
		slog.String("url", r.Request.URL.String()),
		slog.Any("headers", r.Request.Header),
		slog.Int("status", r.StatusCode),
	)
}

func (v *VerboseMiddleware) traceResponse(recorder *http.Response) {
	responseDump, err := httputil.DumpResponse(recorder, true)
	if err != nil {
		slog.Error("failed to dump response", slog.String("error", err.Error()))
		return
	}

	// We need to parse the dump to separate headers and body
	parts := strings.SplitN(string(responseDump), "\r\n\r\n", 2)
	if len(parts) != 2 {
		slog.Log(recorder.Request.Context(), v.level, "HTTP Response", slog.String("dump", string(responseDump)))
		return
	}

	if indented, ok := prettyJSON([]byte(parts[1])); ok {
		slog.Log(recorder.Request.Context(), v.level, "HTTP Response", slog.String("headers", parts[0]), slog.Any("body", indented))
		return
	}

	slog.Log(recorder.Request.Context(), v.level, "HTTP Response", slog.String("dump", string(responseDump)))
}

func prettyJSON(b []byte) (string, bool) {
	var indented bytes.Buffer
	if err := json.Indent(&indented, b, "", "  "); err != nil {
		return "", false
	}
	return indented.String(), true
}
