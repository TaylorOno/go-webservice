package ui

import (
	"embed"
	"html/template"
	"log/slog"
	"net/http"
)

//go:embed templates/*
var templatesFiles embed.FS

var templates *template.Template

func init() {
	templates = template.Must(template.ParseFS(templatesFiles, "templates/*.gohtml"))
}

func RenderTemplate(writer http.ResponseWriter, templateName string, obj any) {
	err := templates.ExecuteTemplate(writer, templateName, obj)
	if err != nil {
		slog.Error("failed to render template: ", "templateName", templateName, "error", err.Error())
		return
	}
}
