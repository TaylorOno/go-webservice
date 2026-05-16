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
	level  slog.Level
	logger *slog.Logger
}

func Verbose(handler slog.Handler, level slog.Level) *VerboseMiddleware {
	return &VerboseMiddleware{
		level:  level,
		logger: slog.New(handler),
	}
}

// RequestLogger prints basic request information to standard output
func (v *VerboseMiddleware) RequestLogger() ClientMiddleware {
	return func(c Doer) Doer {
		return ClientFunc(func(req *http.Request) (*http.Response, error) {
			v.logRequest(req)
			res, err := c.Do(req)
			if err != nil {
				v.logError(req, err)
				return res, err
			}

			v.logResponse(req, res)
			return res, err
		})
	}
}

func (v *VerboseMiddleware) logRequest(r *http.Request) {
	if v.level < -4 {
		v.traceRequest(r)
	}
}

func (v *VerboseMiddleware) traceRequest(req *http.Request) {
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		v.logger.Error("failed to dump request", slog.String("error", err.Error()))
		return
	}

	// We need to parse the dump to separate headers and body
	parts := strings.SplitN(string(requestDump), "\r\n\r\n", 2)
	if len(parts) != 2 {
		v.logger.Log(req.Context(), v.level, "HTTP Request", slog.String("dump", string(requestDump)))
		return
	}

	if indented, ok := prettyJSON([]byte(parts[1])); ok {
		v.logger.Log(req.Context(), v.level, "HTTP Request", slog.String("headers", parts[0]), slog.String("body", indented))
		return
	}

	v.logger.Log(req.Context(), v.level, "HTTP Request", slog.String("dump", string(requestDump)))
}

func (v *VerboseMiddleware) logResponse(req *http.Request, resp *http.Response) {
	switch {
	case v.level < -4:
		v.traceResponse(req, resp)
	case v.level < -0:
		v.debugResponse(req, resp)
	}
}

func (v *VerboseMiddleware) logError(req *http.Request, err error) {
	v.logger.Log(req.Context(), v.level, "HTTP Response",
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.Any("headers", req.Header),
		slog.String("error", err.Error()),
	)
}

func (v *VerboseMiddleware) debugResponse(req *http.Request, resp *http.Response) {
	v.logger.Log(req.Context(), v.level, "HTTP Response",
		slog.String("method", req.Method),
		slog.String("url", req.URL.String()),
		slog.Any("headers", req.Header),
		slog.Int("status", resp.StatusCode),
	)
}

func (v *VerboseMiddleware) traceResponse(req *http.Request, resp *http.Response) {
	responseDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		v.logger.Error("failed to dump response", slog.String("error", err.Error()))
		return
	}

	// We need to parse the dump to separate headers and body
	parts := strings.SplitN(string(responseDump), "\r\n\r\n", 2)
	if len(parts) != 2 {
		v.logger.Log(req.Context(), v.level, "HTTP Response", slog.String("dump", string(responseDump)))
		return
	}

	if indented, ok := prettyJSON([]byte(parts[1])); ok {
		v.logger.Log(req.Context(), v.level, "HTTP Response", slog.String("headers", parts[0]), slog.Any("body", indented))
		return
	}

	v.logger.Log(req.Context(), v.level, "HTTP Response", slog.String("dump", string(responseDump)))
}

func prettyJSON(b []byte) (string, bool) {
	var indented bytes.Buffer
	if err := json.Indent(&indented, b, "", "  "); err != nil {
		return "", false
	}
	return indented.String(), true
}
