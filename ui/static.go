package ui

import (
	"embed"
	"net/http"
)

//go:embed static/*
var staticFiles embed.FS

func ServerStatic(writer http.ResponseWriter, request *http.Request) {
	http.FileServer(http.FS(staticFiles)).ServeHTTP(writer, request)
}
