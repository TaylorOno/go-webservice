package server

import "net/http"

func addRoutes(
	mux *http.ServeMux,
) {
	mux.HandleFunc("/helloworld/", helloWorld)
	mux.Handle("/", http.NotFoundHandler())
}

func helloWorld(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("hello"))
}
