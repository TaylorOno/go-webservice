package ui

import (
	"net/http"

	"github.com/taylorono/go-lib/web"
)

type Mux interface {
	HandleFunc(pattern string, handler http.HandlerFunc)
}

type Handler struct {
}

// NewUIHandler returns a new Handler instance.
func NewUIHandler() *Handler {
	return &Handler{}
}

func (h Handler) Routes(server *web.Server) {
	// route static css/js/asset files
	server.HandleFunc("GET /static/", ServerStatic)

	// redirect base url to home page
	server.HandleFunc("GET /{$}", http.RedirectHandler("/home", http.StatusPermanentRedirect).ServeHTTP)

	// all page routes
	server.HandleFunc("GET /home/{$}", Home)
}
