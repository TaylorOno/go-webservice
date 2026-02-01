package service

import "net/http"

type Mux interface {
	HandleFunc(pattern string, handler http.HandlerFunc)
}

func (s *Service) AddRoutes(mux Mux) {
	mux.HandleFunc("GET /helloworld", helloWorld)
	mux.HandleFunc("/", http.NotFound)
}

func helloWorld(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("hello"))
}
