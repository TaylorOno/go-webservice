package api

import (
	"net/http"

	"github.com/taylorono/go-webservice/internal/service"
)

type Mux interface {
	HandleFunc(pattern string, handler http.HandlerFunc)
}

type GreeterHandler struct {
	Service *service.Service
}

func NewGreeterHandler(service *service.Service) *GreeterHandler {
	return &GreeterHandler{Service: service}
}

func (s *GreeterHandler) Routes(mux Mux) {
	mux.HandleFunc("GET /helloworld", s.helloWorld)
}

func (s *GreeterHandler) helloWorld(w http.ResponseWriter, r *http.Request) {
	greeting := s.Service.SayHello()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(greeting))
}
