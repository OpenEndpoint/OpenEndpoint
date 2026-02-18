// Package dashboard provides a simple embedded web dashboard
package dashboard

import (
	"embed"
	"encoding/json"
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

// ClusterGetter interface for getting cluster info
type ClusterGetter interface {
	GetClusterInfo() interface{}
	GetNodes() interface{}
}

// Handler returns the dashboard HTTP handler
func Handler(clusterGetter ClusterGetter) http.Handler {
	mux := http.NewServeMux()

	// Serve static files
	mux.Handle("/static/", http.FileServer(http.FS(Static)))

	// Serve index
	mux.HandleFunc("/", indexHandler)

	// Serve metrics dashboard
	mux.HandleFunc("/_dashboard/metrics", metricsHandler)

	// Serve cluster dashboard
	mux.HandleFunc("/_dashboard/cluster", clusterHandler(clusterGetter))

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
		Version: "2.0.0",
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
		Version: "2.0.0",
	}

	tmpl.Execute(w, data)
}

func clusterHandler(getter ClusterGetter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check if it's an API request
		if r.URL.Query().Get("format") == "json" || r.Header.Get("Accept") == "application/json" {
			w.Header().Set("Content-Type", "application/json")

			// Get cluster info
			info := getter.GetClusterInfo()
			nodes := getter.GetNodes()

			response := map[string]interface{}{
				"replicationFactor": 3,
				"totalStorage":      int64(0),
				"nodes":            nodes,
			}

			if info != nil {
				if ci, ok := info.(interface{ ReplicationFactor() int }); ok {
					response["replicationFactor"] = ci.ReplicationFactor()
				}
			}

			json.NewEncoder(w).Encode(response)
			return
		}

		// Serve HTML
		tmpl, err := template.ParseFS(Templates, "templates/cluster.html")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		data := struct {
			Title   string
			Version string
		}{
			Title:   "OpenEndpoint Cluster",
			Version: "2.0.0",
		}

		tmpl.Execute(w, data)
	}
}
