package router

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/instantminecraft/server/pkg/config"
	"github.com/instantminecraft/server/pkg/enums"
	"github.com/instantminecraft/server/pkg/manager"
	"github.com/instantminecraft/server/pkg/models"
	"github.com/instantminecraft/server/pkg/utils"
	"golang.org/x/exp/slices"
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

	// check if requested mc version is valid
	if !slices.Contains(config.AvailableVersions, mcVersion) {
		// Requested mc version not valid
		sendError(fmt.Sprintf("mc_version %s not available", mcVersion), w, http.StatusBadRequest)
		return
	}

	serverID := manager.GenerateMcServerID(name)

	preparationChan := manager.AddPreparingServer(serverID)

	// Check if a prepared server with requested mc version exists
	readyContainer, err := manager.GetPreparedMcServerContainerMcVersion(mcVersion)
	if err != nil {
		sendError("Couldn't fetch available server", w, http.StatusInternalServerError)
		return
	}

	if len(readyContainer) > 0 {
		// no need for preparation, we can start a mc server instance instantly
		mcServer, err := manager.StartMcServer(readyContainer[0].ID, name)
		if err != nil {
			sendError("Couldn't start mc server", w, http.StatusInternalServerError)
			return
		}
		data, _ := json.Marshal(mcServer.ToClientJson())
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	} else {
		data, _ := json.Marshal(map[string]interface{}{
			"status":     enums.Preparing.String(),
			"id":         serverID,
			"name":       name,
			"mc_version": mcVersion,
		})
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}

	go func() {
		if len(readyContainer) == 0 {
			// We need to check if the docker image is prepared
			utils.ChanSendString(preparationChan, "Preparing server preparation")
			manager.EnsureImageIsReady(config.ImageWithMcVersion(mcVersion))

			// we need to prepare a server with given mc version
			utils.ChanSendString(preparationChan, "Starting server preparation") //preparationChan <- "Starting server preparation"
			manager.PrepareMcServer(mcVersion)
			utils.ChanSendString(preparationChan, "Waiting for preparation end") //preparationChan <- "Waiting for preparation end"
			manager.WaitForTargetServerPrepared(mcVersion)
		}
		readyContainer, _ := manager.GetPreparedMcServerContainerMcVersion(mcVersion)

		// run a prepared server
		utils.ChanSendString(preparationChan, "Starting server") //preparationChan <- "Starting server"
		_, err := manager.StartMcServer(readyContainer[0].ID, name)
		if err != nil {
			fmt.Println("TODO", err)
			// TODO
			return
		}
		utils.ChanSendString(preparationChan, "Done") //preparationChan <- "Done"
		manager.RemovePreparingServer(serverID)
	}()
}

func serverStartStatus(w http.ResponseWriter, r *http.Request) {
	serverID := mux.Vars(r)["serverid"]
	prepChan := manager.GetPreparingServerChan(serverID)
	if prepChan == nil {
		// channel not found, probably serverID not found
		sendError("Server not found", w, http.StatusNotFound)
		return
	}

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		sendError("Couldn't establish a websocket connection", w, http.StatusBadRequest)
		return
	}

	for {
		message := <-prepChan
		conn.WriteJSON(map[string]string{
			"message": message,
		})
		if message == "Done" {
			break
		}
	}
	conn.Close()
}
