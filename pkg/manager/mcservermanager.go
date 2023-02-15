package manager

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/instantminecraft/server/pkg/api/mcserverapi"
	"github.com/instantminecraft/server/pkg/config"
	"github.com/instantminecraft/server/pkg/models"
	"github.com/instantminecraft/server/pkg/utils"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

var mcServer []models.McServerContainer
var mcServerPreperationWG sync.WaitGroup

const authEnvKey = "auth"

// InitMCServerManagement Setup docker connection and retrieve already running minecraft server container instances
func InitMCServerManagement() {
	possibleUnfinishedPrepContainer, err := ListContainersByNameStart(config.WaitingReadyContainerName)
	if err != nil {
		log.Fatal().Err(err).Msg("Couldn't connect to docker daemon")
		panic(err)
	}

	for _, container := range possibleUnfinishedPrepContainer {
		// check if container is unfinished
		isPaused, err := IsContainerPaused(container.ID)
		if err != nil || !isPaused {
			// container needs to be removed because we can't be sure in what state the container is
			KillContainer(container.ID)
			RemovePortFromUsageList(int(container.Ports[0].PublicPort))
			continue
		}
		// reserve port for this container
		AddPortToUsageList(int(container.Ports[0].PublicPort))
	}

	// check if a container needs to be prepared
	preparedContainer, err := GetPreparedMcServerContainer()
	if err != nil {
		log.Error().Err(err).Msg("Couldn't fetch already prepared containers:")
	} else if len(preparedContainer) == 0 {
		log.Info().Msg("Preparing a mc server in the background...")
		PrepareMcServer()
	} else if len(preparedContainer) > 0 {
		// we need obtain the auth keys
		authKeys := ObtainAuthKeys(preparedContainer)
		MergeAuthKeys(authKeys)
		log.Info().Msg("A mc server is already prepared. Good!")
	}

}

// PrepareMcServer Creates a mc server container, setup the mc world and pause the container for later deployment
func PrepareMcServer() {
	mcServerPreperationWG.Add(1)
	go prepareMcServerSync()
}

func prepareMcServerSync() {
	port := GeneratePort()
	AddPortToUsageList(port)
	authKey := GenerateAuthKeyForMcServer()
	env := []string{"autostart=false", fmt.Sprintf("%s=%s", authEnvKey, authKey)}

	containerName := config.WaitingReadyContainerName
	preparedContainer, err := GetPreparedMcServerContainer()
	if err == nil && len(preparedContainer) > 0 {
		containerName = config.WaitingReadyContainerNr(len(preparedContainer))
	}

	containerID, err := RunContainer(config.LatestImageName, containerName, port, env)
	if err != nil {
		log.Error().Err(err).Msg("Couldn't start preparation docker container. Retrying in 2 seconds...")
		time.Sleep(2 * time.Second)
		RemovePortFromUsageList(port)
		prepareMcServerSync()
		return
	}

	serverStatus, err := mcserverapi.GetServerStatus(port, authKey)
	if serverStatus.Server.Running == false {
		// we need to prepare the minecraft world
		mcserverapi.WaitForMcWorldBootUp(port, authKey)
		// TODO error handling
	}

	// Now we need to pause the container because the mc world needs to stop
	SaveAuthKey(containerID, authKey)
	PauseContainer(containerID)
	log.Info().Msg("A mc server container has been prepared")
	mcServerPreperationWG.Done()
}

// GetPreparedMcServerContainer Returns a list of Container which minecraft world is setup and the container state is paused
func GetPreparedMcServerContainer() ([]types.Container, error) {
	var readyContainer []types.Container
	container, err := ListContainersByNameStart(config.WaitingReadyContainerName)
	if err != nil {
		return nil, err
	}
	for _, curContainer := range container {
		isPaused, err := IsContainerPaused(curContainer.ID)
		if err != nil {
			continue
		}
		if isPaused {
			readyContainer = append(readyContainer, curContainer)
		}
	}
	return readyContainer, nil
}

func PreparedMcServerContainerExists(containerId string) (bool, error) {
	container, err := ListContainersByNameStart(config.WaitingReadyContainerName)
	if err != nil {
		return false, err
	}
	for _, curContainer := range container {
		isPaused, err := IsContainerPaused(curContainer.ID)
		if err != nil {
			continue
		}
		if isPaused && curContainer.ID == containerId {
			return true, nil
		}
	}
	return false, nil
}

func WaitForFinsishedPreparing() {
	mcServerPreperationWG.Wait()
}

func GetRunningMcServer() ([]models.McServerContainer, error) {
	_, err := ListContainer()

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func generateId(serverName string) string {
	return utils.MD5(serverName + utils.RandomString(32))
}

func generateContainerName(serverId string) string {
	return config.ContainerBaseName + serverId
}

func StartMcServer(containerID string, name string) (models.McServerContainer, error) {
	log.Info().Msgf("Looking for prepared container %s...", containerID)
	preparedMcServer, err := GetPreparedMcServerContainer()

	if err != nil {
		return models.McServerContainer{}, err
	}

	var exists bool = false
	var targetContainer types.Container
	for _, preparedServer := range preparedMcServer {
		if preparedServer.ID == containerID {
			exists = true
			targetContainer = preparedServer
			break
		}
	}
	if !exists {
		err := errors.New("Couldn't find prepared container with ID " + containerID)
		log.Error().Err(err).Send()
		return models.McServerContainer{}, err
	}

	log.Info().Msg("Starting Mc Server with container ID " + containerID)

	id := generateId(name)

	err = RenameContainer(containerID, generateContainerName(id))

	if err != nil {
		log.Error().Err(err).Msg("Couldn't rename container")
	}

	err = ResumeContainer(containerID)
	mcVersion := utils.GetMcVersionFromContainer(targetContainer)
	port := utils.GetPortFromContainer(targetContainer)
	return models.McServerContainer{ContainerID: containerID, Name: name, ID: id, Port: port, McVersion: mcVersion, Running: true}, err
}
