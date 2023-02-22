package router

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func sendError(error string, w http.ResponseWriter, status int) {
	data, _ := json.Marshal(map[string]interface{}{
		"error": error,
	})
	w.WriteHeader(status)
	w.Write(data)
}

func rootRoute(w http.ResponseWriter, r *http.Request) {
	data, _ := json.Marshal(map[string]interface{}{
		"server": "InstantMC",
	})
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
