// Package dashboard provides a simple embedded web dashboard
package dashboard

import (
	"embed"
	"html/template"
	"net/http"
)

// Static files embedded
//
//go:embed static/*
var Static embed.FS

// HTML templates
//
//go:embed templates/*.html
var Templates embed.FS

// Handler returns the dashboard HTTP handler
func Handler() http.Handler {
	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/static/", http.FileServer(http.FS(Static)))

	// Serve index
	mux.HandleFunc("/", indexHandler)

	// Serve metrics dashboard
	mux.HandleFunc("/_dashboard/metrics", metricsHandler)

	return mux
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(Templates, "templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	data := struct {
		Title   string
		Version string
	}{
		Title:   "OpenEndpoint Dashboard",
		Version: "0.1.0",
	}

	tmpl.Execute(w, data)
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFS(Templates, "templates/metrics.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	data := struct {
		Title   string
		Version string
	}{
		Title:   "OpenEndpoint Metrics",
		Version: "0.1.0",
	}

	tmpl.Execute(w, data)
}
