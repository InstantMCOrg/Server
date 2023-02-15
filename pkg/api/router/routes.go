package router

import (
	"encoding/json"
	"net/http"
)

func sendError(error string, w http.ResponseWriter, status int) {
	data, _ := json.Marshal(map[string]interface{}{
		"error": error,
	})
	w.WriteHeader(status)
	w.Write(data)
}

func rootRoute(w http.ResponseWriter, r *http.Request) {
	data, _ := json.Marshal(map[string]interface{}{
		"server": "InstantMinecraft",
	})
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
