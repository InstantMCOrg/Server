package manager

import (
	"github.com/docker/docker/api/types"
	"github.com/instantminecraft/server/pkg/api/mcserverapi"
	"github.com/instantminecraft/server/pkg/config"
	"github.com/instantminecraft/server/pkg/models"
	"github.com/rs/zerolog/log"
	"sync"
	"time"
)

var mcServer []models.MCServer
var mcServerPreperationWG sync.WaitGroup

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
	preparedContainer, err := GetPreparedMcServer()
	if err != nil {
		log.Error().Err(err).Msg("Couldn't fetch already prepared containers:")
	} else if len(preparedContainer) == 0 {
		log.Info().Msg("Preparing a mc server in the background...")
		PrepareMcServer()
	} else if len(preparedContainer) > 0 {
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
	noAutoStartEnv := []string{"autostart=false"}

	containerName := config.WaitingReadyContainerName
	preparedContainer, err := GetPreparedMcServer()
	if err == nil && len(preparedContainer) > 0 {
		containerName = config.WaitingReadyContainerNr(len(preparedContainer))
	}
	containerID, err := RunContainer(config.LatestImageName, containerName, port, noAutoStartEnv)
	if err != nil {
		log.Error().Err(err).Msg("Couldn't start preparation docker container. Retrying in 2 seconds...")
		time.Sleep(2 * time.Second)
		RemovePortFromUsageList(port)
		prepareMcServerSync()
		return
	}

	serverStatus, err := mcserverapi.GetServerStatus(port)
	if serverStatus.Server.Running == false {
		// we need to prepare the minecraft world
		mcserverapi.WaitForMcWorldBootUp(port)
		// TODO error handling
	}

	// Now we need to pause the container because the mc world needs to stop
	PauseContainer(containerID)
	log.Info().Msg("A mc server container has been prepared")
	mcServerPreperationWG.Done()
}

// GetPreparedMcServer Returns a list of Container which minecraft world is setup and the container state is paused
func GetPreparedMcServer() ([]types.Container, error) {
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

func WaitForFinsishedPreparing() {
	mcServerPreperationWG.Wait()
}

func StartMcServer() error {
	log.Info().Msg("Looking for prepared Container...")
	preparedMcServer, err := GetPreparedMcServer()

	if err != nil {
		return err
	}

	if len(preparedMcServer) == 0 {
		log.Info().Msg("Preparing Container...")
		PrepareMcServer()
		WaitForFinsishedPreparing()
		StartMcServer()
		return nil
	}

	log.Info().Msg("Resuming container...")
	err = ResumeContainer(preparedMcServer[0].ID)
	return err
}
