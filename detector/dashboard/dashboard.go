package dashboard

import (
	"encoding/json"
	"net/http"

	"detector-app/metrics"
)

// Start launches the metrics dashboard HTTP server on port 8081.
// Serves the static UI at / and JSON metrics at /metrics.
func Start() {
	mux := http.NewServeMux()

	mux.Handle("/", http.FileServer(http.Dir("./dashboard/static")))

	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(metrics.Get())
	})

	http.ListenAndServe(":8081", mux)
}
