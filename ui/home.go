package ui

import (
	"net/http"
)

type HomePage struct{}

func Home(w http.ResponseWriter, r *http.Request) {
	RenderTemplate(w, "home", HomePage{})
}
