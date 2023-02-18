package manager

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/instantminecraft/server/pkg/api/mcserverapi"
	"github.com/instantminecraft/server/pkg/config"
	"github.com/instantminecraft/server/pkg/enums"
	"github.com/instantminecraft/server/pkg/models"
	"github.com/instantminecraft/server/pkg/utils"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var mcServer []models.McServerContainer
var mcServerPreparationWG sync.WaitGroup

// mcServerVersionPreparationWg defines a waitgroup for a mc version
var mcServerVersionPreparationWG = map[string]*sync.WaitGroup{}

// the variable has the following structure: map[serverID]preperation status channel
var preparingMcContainer = map[string]chan string{}

const authEnvKey = "auth"

// InitMCServerManagement Setup docker connection and retrieve already running minecraft server container instances
func InitMCServerManagement() {
	containerList, err := ListContainer()
	var preparedContainer []types.Container
	if err != nil {
		log.Fatal().Err(err).Msg("Couldn't connect to docker daemon")
		panic(err)
	}

	for _, container := range containerList {
		if IsContainerPreparationServer(container) {
			// check if container is unfinished
			isPaused, err := IsContainerPaused(container.ID)
			if err != nil || !isPaused {
				// container needs to be removed because we can't be sure in what state the container is
				KillContainer(container.ID)
				RemovePortFromUsageList(int(container.Ports[0].PublicPort))
				continue
			}
			preparedContainer = append(preparedContainer, container)
		}

		// reserve port for this container
		AddPortToUsageList(int(container.Ports[0].PublicPort))
	}

	// check if a container needs to be prepared
	if len(preparedContainer) == 0 {
		log.Info().Msg("Preparing a mc server in the background...")
		PrepareMcServer(config.LatestMcVersion, models.McServerPreparationConfig{})
	} else if len(preparedContainer) > 0 {
		// we need obtain the auth keys
		authKeys := ObtainAuthKeys(preparedContainer)
		MergeAuthKeys(authKeys)
		log.Info().Msg("A mc server is already prepared. Good!")
	}

}

// PrepareMcServer Creates a mc server container, setup the mc world and pause the container for later deployment
// World mount path has the following system: `<current dir>/worlds/<port of container>`
func PrepareMcServer(mcVersion string, preparationConfig models.McServerPreparationConfig) {
	mcServerPreparationWG.Add(1)
	wg, ok := mcServerVersionPreparationWG[mcVersion]
	if !ok {
		var newWaitGroup sync.WaitGroup
		wg = &newWaitGroup
		mcServerVersionPreparationWG[mcVersion] = wg
	}
	wg.Add(1)
	go prepareMcServerSync(mcVersion, preparationConfig)
}

func prepareMcServerSync(mcVersion string, preparationConfig models.McServerPreparationConfig) {
	var port int
	if preparationConfig.Port == 0 {
		port = GeneratePort()
	} else {
		port = preparationConfig.Port
	}
	AddPortToUsageList(port)
	var authKey string
	if preparationConfig.AuthKey == "" {
		authKey = GenerateAuthKeyForMcServer()
	} else {
		authKey = preparationConfig.AuthKey
	}

	env := []string{"autostart=false", fmt.Sprintf("%s=%s", authEnvKey, authKey)}

	containerName := config.WaitingReadyContainerName
	preparedContainer, err := GetPreparedMcServerContainer()
	if err == nil && len(preparedContainer) > 0 {
		containerName = config.WaitingReadyContainerNr(len(preparedContainer))
	}

	currentPath, _ := os.Getwd()
	targetWorldMountPath := filepath.Join(currentPath, "worlds", fmt.Sprintf("%d", port))

	containerID, err := RunContainer(config.ImageWithMcVersion(mcVersion), containerName, port, env, targetWorldMountPath)
	if err != nil {
		log.Error().Err(err).Msg("Couldn't start preparation docker container. Retrying in 2 seconds...")
		time.Sleep(2 * time.Second)
		RemovePortFromUsageList(port)
		prepareMcServerSync(mcVersion, preparationConfig)
		return
	}

	if preparationConfig.CoreBootUpWG != nil {
		preparationConfig.CoreBootUpWG.Done()
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
	log.Info().Msgf("A mc %s server container has been prepared", mcVersion)
	mcServerPreparationWG.Done()
	mcServerVersionPreparationWG[mcVersion].Done()
}

func IsContainerPreparationServer(container types.Container) bool {
	return len(container.Names) > 0 && strings.HasPrefix(container.Names[0], "/"+config.WaitingReadyContainerName)
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

// GetPreparedMcServerContainerMcVersion Returns a list of prepared container with a specific mc version
func GetPreparedMcServerContainerMcVersion(targetMcVersion string) ([]types.Container, error) {
	readyContainer, err := GetPreparedMcServerContainer()
	if err != nil {
		return nil, err
	}

	var containerWithTargetMcVersion []types.Container

	for _, container := range readyContainer {
		mcVersion := utils.GetMcVersionFromContainer(container)
		if mcVersion == targetMcVersion {
			containerWithTargetMcVersion = append(containerWithTargetMcVersion, container)
		}
	}

	return containerWithTargetMcVersion, nil
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

func WaitForFinishedPreparing() {
	mcServerPreparationWG.Wait()
}

func WaitForTargetServerPrepared(mcVersion string) {
	if wg, ok := mcServerVersionPreparationWG[mcVersion]; ok {
		wg.Wait()
	}
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

func GenerateMcServerID(serverName string) string {
	return generateId(serverName)
}

func AddPreparingServer(serverID string) chan string {
	preparingMcContainer[serverID] = make(chan string)
	return preparingMcContainer[serverID]
}

func RemovePreparingServer(serverID string) {
	newMap := map[string]chan string{}
	for k, v := range preparingMcContainer {
		if k != serverID {
			newMap[k] = v
		}
	}

	preparingMcContainer = newMap
}

func GetPreparingServerChan(serverID string) chan string {
	return preparingMcContainer[serverID]
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
	return models.McServerContainer{ContainerID: containerID, Name: name, ID: id, Port: port, McVersion: mcVersion, Status: enums.Running}, err
}
