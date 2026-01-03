package service

import "net/http"

type httpServer interface {
	AddRoute(pattern string, handler http.HandlerFunc)
}

func (s *Service) AddRoutes(httpServer httpServer) {
	httpServer.AddRoute("GET /helloworld", helloWorld)
	httpServer.AddRoute("/", http.NotFound)
}

func helloWorld(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("hello"))
}
