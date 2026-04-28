package dashboard

import (
	"encoding/json"
	"net/http"

	"detector-app/metrics"
)

func Start() {
	http.Handle("/", http.FileServer(http.Dir("./dashboard/static")))

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(metrics.Get())
	})

	http.ListenAndServe(":8081", nil)
}
