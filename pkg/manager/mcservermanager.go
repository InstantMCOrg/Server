package manager

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/instantminecraft/server/pkg/api/mcserverapi"
	"github.com/instantminecraft/server/pkg/config"
	"github.com/instantminecraft/server/pkg/db"
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
const ramEnvKey = "ram"

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

	// Now we need to check for saved servers in the db
	savedServer, err := db.GetSavedMcServer()
	if err != nil {
		log.Fatal().Err(err).Msg("Couldn't fetch saved mc servers in db")
	}
	alreadyRunningServer, err := GetRunningMcServer()
	if err != nil {
		log.Fatal().Err(err).Msg("Couldn't fetch already running server")
	}

	for _, server := range savedServer {
		// check if the current server is already running
		targetServerID := server.ServerID
		exists := false
		for _, runningServer := range alreadyRunningServer {
			if runningServer.ServerID == targetServerID {
				exists = true
				// we need to update the db for that server, specifically the containerID
				if err := db.UpdateServerContainerID(&server, runningServer.ContainerID); err != nil {
					log.Error().Err(err).Msgf("Couldn't update db entry for containerID for server %s", targetServerID)
				}
				break
			}
		}
		if exists {
			log.Info().Msgf("☑ Mc server %s is already running", targetServerID)
		} else {
			log.Info().Msgf("☐ Mc server %s is starting...", targetServerID)
			var coreBootUpWaitGroup sync.WaitGroup
			coreBootUpWaitGroup.Add(1)
			PrepareMcServer(server.McVersion, models.McServerPreparationConfig{
				Port:         server.Port,
				RamSizeMB:    server.RamSizeMB,
				CoreBootUpWG: &coreBootUpWaitGroup,
				ServerID:     server.ServerID,
				AutoDeploy:   true,
			})
			go func() {
				coreBootUpWaitGroup.Wait()
				log.Info().Msgf("☑ Mc server %s started successfully", targetServerID)
			}()
		}
	}

}

// PrepareMcServer Creates a mc server container, setup the mc world and pause the container for later deployment
// World mount path has the following system: `<current dir>/worlds/<port of container>`
// If models.McServerPreparationConfig CoreBootUpWG is not nil, you need to call .Add(1) before calling PrepareMcServer
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
	var targetRamSize = config.DefaultRamSize
	if preparationConfig.RamSizeMB != 0 {
		targetRamSize = preparationConfig.RamSizeMB
	}
	env = append(env, fmt.Sprintf("ram=%d", targetRamSize))

	containerName := config.WaitingReadyContainerName
	if preparationConfig.ServerID != "" {
		containerName = generateContainerName(preparationConfig.ServerID)
	} else {
		// normal container preparation
		preparedContainer, err := GetPreparedMcServerContainer()
		if err == nil && len(preparedContainer) > 0 {
			containerName = config.WaitingReadyContainerNr(len(preparedContainer))
		}
	}

	currentPath, _ := os.Getwd()
	targetWorldMountPath := filepath.Join(currentPath, config.DataDir, "worlds", fmt.Sprintf("%d", port))

	containerID, err := RunContainer(config.ImageWithMcVersion(mcVersion), containerName, port, env, targetWorldMountPath, targetRamSize)
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
	if !preparationConfig.AutoDeploy {
		PauseContainer(containerID)
		log.Info().Msgf("A mc %s server container has been prepared", mcVersion)
	}

	mcServerPreparationWG.Done()
	mcServerVersionPreparationWG[mcVersion].Done()
}

func IsContainerPreparationServer(container types.Container) bool {
	return len(container.Names) > 0 && strings.HasPrefix(container.Names[0], "/"+config.WaitingReadyContainerName)
}

// GetServerIDFromContainer Returns the serverID or an empty string ("") if the container is not an mc server
func GetServerIDFromContainer(container types.Container) string {
	baseName := "/" + config.ContainerBaseName
	if len(container.Names) > 0 && strings.HasPrefix(container.Names[0], baseName) {
		return strings.ReplaceAll(container.Names[0], baseName, "")
	}
	return ""
}

func GetMcServerContainer(searchConfig models.McContainerSearchConfig) ([]types.Container, error) {
	var targetContainer []types.Container

	searchForPreparedContainer := searchConfig.Status == enums.Prepared
	var container []types.Container
	var err error
	if searchForPreparedContainer {
		container, err = ListContainersByNameStart(config.WaitingReadyContainerName)
	} else {
		container, err = ListContainersByNameStart(config.ContainerBaseName)
	}
	if err != nil {
		return nil, err
	}

	// now we need to filter
	for _, curContainer := range container {
		if searchForPreparedContainer {
			// we have to check if the container is paused. If not the container is not ready prepared
			isPaused, err := IsContainerPaused(curContainer.ID)
			if err != nil || !isPaused {
				// throw it away
				continue
			}
		} else {
			// Prepared unused container must go out
			if IsContainerPreparationServer(curContainer) {
				continue
			}
		}
		if searchConfig.McVersion != "" {
			// we need to match the mc version
			mcVersion := utils.GetMcVersionFromContainer(curContainer)
			if mcVersion != searchConfig.McVersion {
				// Mc version doesn't match
				continue
			}
		}
		if searchConfig.RamSizeMB != 0 {
			// we need to match the target ram size
			ramSize, err := GetContainerRamSizeEnv(curContainer.ID)
			if err != nil || ramSize != searchConfig.RamSizeMB {
				continue
			}
		}

		targetContainer = append(targetContainer, curContainer)
	}

	return targetContainer, nil
}

// GetPreparedMcServerContainer Returns a list of Container which minecraft world is setup and the container state is paused
// Deprecated: Use GetMcServerContainer with config instead
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
// Deprecated: Use GetMcServerContainer with config instead
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
	mcServer, err := GetMcServerContainer(models.McContainerSearchConfig{
		Status: enums.Running,
	})
	if err != nil {
		return nil, err
	}
	var result []models.McServerContainer

	for _, container := range mcServer {
		serverID := GetServerIDFromContainer(container)
		if serverID == "" {
			// not a valid mc server container
			continue
		}
		serverData, err := db.GetMcServerData(serverID)
		if err != nil {
			// server is not in db
			continue
		}
		// currently the containerID in serverData got replaced by the dbs value. We need to fix that
		serverData.ContainerID = container.ID
		result = append(result, *serverData.Self())
	}

	return result, nil
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
	ram, _ := GetContainerRamSizeEnv(containerID)
	return models.McServerContainer{ContainerID: containerID, Name: name, ServerID: id, Port: port, McVersion: mcVersion, Status: enums.Running, RamSizeMB: ram}, err
}
