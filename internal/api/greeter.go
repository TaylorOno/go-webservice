package api

import (
	"log/slog"
	"net/http"

	"github.com/taylorono/go-lib/logging"
	"github.com/taylorono/go-lib/web"
	"github.com/taylorono/go-webservice/internal/service"
)

var (
	statusGetJokeFailed = logging.NewStatus("api", "failed to get joke")
	statusParseError    = logging.NewStatus("api", "parse error")
)

type Mux interface {
	HandleFunc(pattern string, handler http.HandlerFunc)
}

// GreeterHandler wraps a service instance and provides HTTP handlers for greeting operations.
type GreeterHandler struct {
	Service *service.Greeter
}

// NewGreeterHandler returns a new GreeterHandler instance.
func NewGreeterHandler(service *service.Greeter) *GreeterHandler {
	return &GreeterHandler{Service: service}
}

// Routes registers HTTP handlers for greeting operations.
func (s *GreeterHandler) Routes(mux Mux) {
	mux.HandleFunc("GET /helloworld", s.helloWorld)
	mux.HandleFunc("POST /helloworld", s.helloUser)
	mux.HandleFunc("GET /dailyjoke", s.helloWithJoke)
}

func (s *GreeterHandler) helloWorld(w http.ResponseWriter, r *http.Request) {
	greeting := s.Service.SayHello()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(greeting))
}

func (s *GreeterHandler) helloUser(w http.ResponseWriter, r *http.Request) {
	ctx := logging.WithLogContext(r.Context(), slog.String("api", "/helloworld"))

	req, err := web.Decode[GreetRequest](r)
	if err != nil {
		slog.ErrorContext(ctx, "failed to parse user name", logging.Metric(statusParseError), "error", err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	greeting := s.Service.SayHelloUser(req.Name)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(greeting))
}

func (s *GreeterHandler) helloWithJoke(w http.ResponseWriter, r *http.Request) {
	ctx := logging.WithLogContext(r.Context(), slog.String("api", "/dailyjoke"))

	joke, err := s.Service.SayMorningJokes(r.Context())
	if err != nil {
		slog.ErrorContext(ctx, "failed to get joke", logging.Metric(statusGetJokeFailed), "error", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(joke))
}

type GreetRequest struct {
	Name string `json:"name"`
}
