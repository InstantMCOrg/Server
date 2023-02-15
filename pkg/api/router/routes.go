package router

import (
	"encoding/json"
	"github.com/instantminecraft/server/pkg/manager"
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

func getPreparedServer(w http.ResponseWriter, r *http.Request) {
	data, _ := json.Marshal(map[string]interface{}{
		"server": map[string]interface{}{
			"running": false,
		},
	})
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func startServer(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	if name == "" {
		sendError("Please provide the field \"name\"", w, http.StatusBadRequest)
		return
	}
	mcVersion := r.FormValue("mc_version")
	if mcVersion == "" {
		sendError("Please provide the field \"mc_version\"", w, http.StatusBadRequest)
		return
	}

	// TODO handle mc version request

	alreadyPreparedContainerId := r.FormValue("container_id")
	useAlreadyPreparedServer := alreadyPreparedContainerId != ""

	if useAlreadyPreparedServer {
		// Check if prepared server exists
		preparedContainerExists, err := manager.PreparedMcServerContainerExists(alreadyPreparedContainerId)
		if err != nil {
			sendError("Couldn't fetch prepared container", w, http.StatusInternalServerError)
			return
		} else if !preparedContainerExists {
			sendError("Container with given ID doesn't exist", w, http.StatusBadRequest)
			return
		}
	} else {
		// We need to prepare a server
		// TODO
	}

	mcServer, err := manager.StartMcServer(alreadyPreparedContainerId, name)
	if err != nil {
		sendError("Couldn't start mc server", w, http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(mcServer)

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
