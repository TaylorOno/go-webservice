package api

import (
	"net/http"

	"github.com/taylorono/go-webservice/internal/framework/web"
	"github.com/taylorono/go-webservice/internal/service"
)

type Mux interface {
	HandleFunc(pattern string, handler http.HandlerFunc)
}

// GreeterHandler wraps a service instance and provides HTTP handlers for greeting operations.
type GreeterHandler struct {
	Service *service.Service
}

// NewGreeterHandler returns a new GreeterHandler instance.
func NewGreeterHandler(service *service.Service) *GreeterHandler {
	return &GreeterHandler{Service: service}
}

// Routes registers HTTP handlers for greeting operations.
func (s *GreeterHandler) Routes(mux Mux) {
	mux.HandleFunc("GET /helloworld", s.helloWorld)
	mux.HandleFunc("POST /helloworld", s.helloUser)
}

func (s *GreeterHandler) helloWorld(w http.ResponseWriter, r *http.Request) {
	greeting := s.Service.SayHello()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(greeting))
}

func (s *GreeterHandler) helloUser(w http.ResponseWriter, r *http.Request) {
	req, err := web.Decode[GreetRequest](r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	greeting := s.Service.SayHelloUser(req.Name)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(greeting))
}

type GreetRequest struct {
	Name string `json:"name"`
}
