package api

import (
	"net/http"

	"github.com/taylorono/go-webservice/internal/service"
)

type ServiceHandlers struct {
	Service *service.Service
}

func NewServiceHandlers(service *service.Service) *ServiceHandlers {
	return &ServiceHandlers{Service: service}
}

func (s *ServiceHandlers) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /helloworld", s.helloWorld)
}

func (s *ServiceHandlers) helloWorld(writer http.ResponseWriter, request *http.Request) {
	greeting := s.Service.SayHello()

	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(greeting))
}
