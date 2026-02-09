package api

import (
	"net/http"

	"github.com/taylorono/go-webservice/internal/service"
)

type Mux interface {
	HandleFunc(pattern string, handler http.HandlerFunc)
}

type ServiceHandlers struct {
	Service *service.Service
}

func NewServiceHandlers(service *service.Service) *ServiceHandlers {
	return &ServiceHandlers{Service: service}
}

func (s *ServiceHandlers) Routes(mux Mux) {
	mux.HandleFunc("GET /helloworld", s.helloWorld)
}

func (s *ServiceHandlers) helloWorld(w http.ResponseWriter, r *http.Request) {
	greeting := s.Service.SayHello()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(greeting))
}
