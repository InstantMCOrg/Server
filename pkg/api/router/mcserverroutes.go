package router

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/instantmc/server/pkg/api/mcserverapi"
	"github.com/instantmc/server/pkg/config"
	"github.com/instantmc/server/pkg/db"
	"github.com/instantmc/server/pkg/enums"
	"github.com/instantmc/server/pkg/manager"
	"github.com/instantmc/server/pkg/models"
	"github.com/instantmc/server/pkg/utils"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"
	"net/http"
	"strconv"
	"sync"
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
		ramSize, _ := manager.GetContainerRamSizeEnv(curContainer.ID)
		result = append(result, models.PreparedContainer{Number: nr, McVersion: mcVersion, RamSizeMB: ramSize})
	}

	data, _ := json.Marshal(map[string]interface{}{
		"prepared_server": result,
	})
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func getServer(w http.ResponseWriter, r *http.Request) {
	server, err := manager.GetRunningMcServer()
	if err != nil {
		sendError("Couldn't fetch running mc server", w, http.StatusInternalServerError)
		return
	}

	serverData := []interface{}{}
	for _, curServer := range server {
		serverData = append(serverData, curServer.ToClientJson())
	}

	data, _ := json.Marshal(map[string]interface{}{
		"server": serverData,
	})
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func deleteServer(w http.ResponseWriter, r *http.Request) {
	serverID := mux.Vars(r)["serverid"]
	// we need to check if the server exists
	mcServerData, err := db.GetMcServerData(serverID)
	if err != nil {
		sendError("Server with given ID doesn't exist", w, http.StatusNotFound)
		return
	}
	// stop container if it's running
	runningMcServer, err := manager.GetRunningMcServer()
	if err != nil {
		sendError("Couldn't get running server", w, http.StatusInternalServerError)
		return
	}

	for _, container := range runningMcServer {
		if container.ServerID == mcServerData.ServerID {
			// kill the container
			err := manager.KillContainer(container.ContainerID)
			if err != nil {
				sendError("Couldn't stop server", w, http.StatusInternalServerError)
				log.Error().Err(err).Msgf("Couldn't stop container %s (Server %s)", container.ContainerID, container.ServerID)
				return
			}
			break
		}
	}

	// now we need to delete the mc world volume
	if err := manager.DeleteMcWorld(mcServerData.Port); err != nil {
		sendError("Couldn't delete mc world", w, http.StatusInternalServerError)
		log.Error().Err(err).Msgf("Couldn't delete mc world with port %d", mcServerData.Port)
		return
	}

	// we need to make the port available again
	manager.RemovePortFromUsageList(mcServerData.Port)

	// finally we delete the db entry
	if err := db.DeleteServer(&mcServerData); err != nil {
		sendError("Couldn't delete server from db", w, http.StatusInternalServerError)
		return
	}

	data, _ := json.Marshal(map[string]interface{}{})
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
	targetRamSizeRaw := r.FormValue("ram") // Optional
	var targetRamSize int = config.DefaultRamSize
	if targetRamSizeRaw != "" {
		var err error
		targetRamSize, err = strconv.Atoi(targetRamSizeRaw)
		if err != nil {
			sendError("Couldn't parse field \"ram\"", w, http.StatusBadRequest)
			return
		} else if targetRamSize > config.MaximumRamPerInstance {
			sendError(fmt.Sprintf("Requested ram exceeds maximum allowed ram size (%dmb)", config.MaximumRamPerInstance), w, http.StatusBadRequest)
			return
		}
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
	readyContainer, err := manager.GetMcServerContainer(models.McContainerSearchConfig{
		McVersion: mcVersion,
		RamSizeMB: targetRamSize,
		Status:    enums.Prepared,
	})
	if err != nil {
		sendError("Couldn't fetch available server", w, http.StatusInternalServerError)
		return
	}
	user, err := getCurrentUser(r)
	if err != nil {
		sendError("Couldn't fetch current user", w, http.StatusInternalServerError)
		return
	}

	if len(readyContainer) > 0 {
		// no need for preparation, we can start a mc server instance instantly
		mcServer, err := manager.StartMcServer(readyContainer[0].ID, name)
		authKey := manager.GetAuthKeyForMcServer(readyContainer[0].ID)
		mcserverapi.SendMessage(mcServer.Port, authKey, "Server wake up successful")
		if err != nil {
			sendError("Couldn't start mc server", w, http.StatusInternalServerError)
			return
		}
		if err := db.AddMcServerContainer(&user, &mcServer); err != nil {
			sendError("Couldn't add mc server to database", w, http.StatusInternalServerError)
			return
		}

		data, _ := json.Marshal(mcServer.ToClientJson())
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	} else {
		data, _ := json.Marshal(map[string]interface{}{
			"status":      enums.Preparing.String(),
			"server_id":   serverID,
			"name":        name,
			"ram_size_mb": targetRamSize,
			"mc_version":  mcVersion,
		})
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}

	go func() {

		// We need to check if the docker image is prepared
		utils.ChanSendString(preparationChan, "Preparing server preparation")
		manager.EnsureImageIsReady(config.ImageWithMcVersion(mcVersion))

		// we need to prepare a server with given mc version
		utils.ChanSendString(preparationChan, "Starting server preparation")

		port := manager.GeneratePort()
		authKey := manager.GenerateAuthKeyForMcServer()

		coreBootUpWaitGroup := sync.WaitGroup{}
		coreBootUpWaitGroup.Add(1)
		manager.PrepareMcServer(mcVersion, models.McServerPreparationConfig{
			Port:         port,
			AuthKey:      authKey,
			CoreBootUpWG: &coreBootUpWaitGroup,
			RamSizeMB:    targetRamSize,
			ServerID:     serverID,
			AutoDeploy:   true,
		})
		coreBootUpWaitGroup.Wait()
		worldGenerationChan, err := mcserverapi.GetWorldGenerationChan(port, authKey)
		if err != nil {
			log.Warn().Err(err).Msgf("Couldn't connect to world generation ws of container on port %d", port)
		} else {
			for {
				worldGenerationPercent := <-worldGenerationChan
				utils.ChanSendString(preparationChan, fmt.Sprintf("Preparing world %d%%", worldGenerationPercent))
				if worldGenerationPercent == 100 {
					break
				}
			}
		}

		utils.ChanSendString(preparationChan, "Waiting for preparation end")
		manager.WaitForTargetServerPrepared(mcVersion) // TODO should be migrated to dedicated sync.WaitGroup
		mcServer, err := manager.GetMcServerContainerByServerID(serverID, name)
		if err != nil {
			utils.ChanSendString(preparationChan, "Couldn't end preparation")
		}

		if err := db.AddMcServerContainer(&user, &mcServer); err != nil {
			utils.ChanSendString(preparationChan, "Couldn't add server to database")
			return
		}

		utils.ChanSendString(preparationChan, "Done")
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

func serverStats(w http.ResponseWriter, r *http.Request) {
	serverID := mux.Vars(r)["serverid"]

	// check if server exist
	serverData, err := db.GetMcServerData(serverID)
	if err != nil {
		sendError("Server with given ID not found", w, http.StatusNotFound)
		return
	}

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		sendError("Couldn't establish a websocket connection", w, http.StatusBadRequest)
		return
	}

	jsonDataChan := make(chan string)
	go func() {
		for {
			jsonMessage := <-jsonDataChan
			conn.WriteMessage(websocket.TextMessage, []byte(jsonMessage))
		}
	}()
	err = manager.SubscribeToContainerStats(serverData.ContainerID, &jsonDataChan)
	conn.WriteJSON(map[string]interface{}{
		"message": "An error occurred: " + err.Error(),
	})
	conn.Close()
}
