package router

import (
	"encoding/json"
	"github.com/instantminecraft/server/pkg/manager"
	"github.com/instantminecraft/server/pkg/models"
	"github.com/instantminecraft/server/pkg/utils"
	"net/http"
)

func getPreparedServer(w http.ResponseWriter, r *http.Request) {
	container, err := manager.GetPreparedMcServerContainer()
	if err != nil {
		sendError("Couldn't fetch prepared server", w, http.StatusInternalServerError)
		return
	}

	var result = []models.PreparedContainer{}

	for nr, curContainer := range container {
		mcVersion := utils.GetMcVersionFromContainer(curContainer)
		result = append(result, models.PreparedContainer{Number: nr, McVersion: mcVersion})
	}

	data, _ := json.Marshal(map[string]interface{}{
		"prepared_server": result,
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
