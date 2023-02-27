package manager

import (
	"bufio"
	"encoding/json"
	"github.com/docker/docker/api/types/mount"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/docker/distribution/context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/instantmc/server/pkg/config"
	"github.com/rs/zerolog/log"
)

var ctx = context.Background()
var cli *client.Client

// InitDockerSystem establishes a connection with the docker daemon. Must be called before any other operations in this file
func InitDockerSystem() {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	cli = dockerClient

	ensureMCServerImageIsReady()
}

func ensureMCServerImageIsReady() {
	EnsureImageIsReady(config.LatestImageName)

	// TODO: implement progress bar through an chan
	log.Info().Msg("Mc server image is ready")
}

// EnsureImageIsReady blocks the execution until the requested docker image is downloaded and ready
func EnsureImageIsReady(imageName string) {
	pullResp, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		log.Error().Err(err).Msgf("Oh oh! An error occurred while downloading the newest %s image. Retrying in 2 seconds...", imageName)
		time.Sleep(2 * time.Second)
		EnsureImageIsReady(imageName)
		return
	}
	reader := bufio.NewReader(pullResp)
	defer pullResp.Close()
	// now we need to wait until the pull finished

	for {
		_, err := reader.ReadBytes('\n')
		if err != nil {
			break
		}

		//log.Print(string(line))
		//log.Print("\n")
	}
}

func ListContainer() ([]types.Container, error) {
	return cli.ContainerList(ctx, types.ContainerListOptions{})
}

func ListContainersByNameStart(namePrefix string) ([]types.Container, error) {
	allContainer, err := ListContainer()
	if err != nil {
		return nil, err
	}

	var container []types.Container

	for _, curContainer := range allContainer {
		if len(curContainer.Names) > 0 && strings.HasPrefix(curContainer.Names[0], "/"+namePrefix) {
			container = append(container, curContainer)
		}
	}
	return container, nil
}

// RunContainer Attempts to run a container with given arguments
// Returns container ID as a string and nil if successful
// Otherwise an empty string and an error
func RunContainer(imageName string, containerName string, containerPort int, env []string, worldMountPath string, ramSizeMB int) (string, error) {
	port := strconv.Itoa(config.McServerProxyPort) + "/tcp"
	ramSizeInBytes := int64(float64(ramSizeMB) * math.Pow(10, 6))

	// we need to create the mount directory first
	err := CreateMcWorld(containerPort)
	if err != nil {
		return "", err
	}
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		ExposedPorts: nat.PortSet{
			nat.Port(port): {},
		},
		Env: env,
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(port): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: strconv.Itoa(containerPort)}},
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: worldMountPath,
				Target: "/server/world",
			},
		},
		Resources: container.Resources{
			Memory: ramSizeInBytes,
		},
	}, nil, nil, containerName)

	if err != nil {
		return "", err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}

func PauseContainer(containerID string) error {
	return cli.ContainerPause(ctx, containerID)
}

func ResumeContainer(containerID string) error {
	return cli.ContainerUnpause(ctx, containerID)
}

func RenameContainer(containerID string, name string) error {
	return cli.ContainerRename(ctx, containerID, name)
}

func ObtainAuthKeys(container []types.Container) map[string]string {
	authKeys := map[string]string{}

	for _, curContainer := range container {
		stats, err := GetContainerStats(curContainer.ID)
		if err != nil {
			continue
		}
		for _, curEnv := range stats.Config.Env {
			if strings.HasPrefix(curEnv, authEnvKey) {
				authKey := strings.Split(curEnv, "=")[1]
				authKeys[curContainer.ID] = authKey
			}
		}
	}

	return authKeys
}

func GetContainerStats(containerID string) (types.ContainerJSON, error) {
	return cli.ContainerInspect(ctx, containerID)
}

func GetContainerRamSizeEnv(containerID string) (int, error) {
	stats, err := GetContainerStats(containerID)
	if err != nil {
		return 0, err
	}
	for _, curEnv := range stats.Config.Env {
		if strings.HasPrefix(curEnv, ramEnvKey) {
			ramRaw := strings.Split(curEnv, "=")[1]
			return strconv.Atoi(ramRaw)
		}
	}
	return config.DefaultRamSize, nil
}

func IsContainerPaused(containerID string) (bool, error) {
	containerStats, err := GetContainerStats(containerID)
	if err != nil {
		return false, err
	}
	return containerStats.State.Paused, nil
}

func KillContainer(containerID string) error {
	return cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true, RemoveVolumes: true})
}

func SubscribeToContainerStats(containerID string, jsonStats *chan string) error {
	containerStats, err := cli.ContainerStats(ctx, containerID, true)
	if err != nil {
		return err
	}
	bufReader := bufio.NewReader(containerStats.Body)
	for {
		bytes, err := bufReader.ReadBytes('\n')
		if err != nil {
			break
		}
		jsonData := map[string]interface{}{}

		err = json.Unmarshal(bytes, &jsonData)
		if err != nil {
			return err
		}

		containerCpuUsage := jsonData["cpu_stats"].(map[string]interface{})["cpu_usage"].(map[string]interface{})["total_usage"].(float64)
		systemCpuUsage := jsonData["cpu_stats"].(map[string]interface{})["system_cpu_usage"].(float64)
		var percentCpuUsage float64
		if containerCpuUsage == 0 || systemCpuUsage == 0 {
			percentCpuUsage = 0
		} else {
			percentCpuUsage = containerCpuUsage / systemCpuUsage
		}

		memoryUsage := jsonData["memory_stats"].(map[string]interface{})["usage"].(float64)        // bytes
		memoryMaxUsage := jsonData["memory_stats"].(map[string]interface{})["max_usage"].(float64) // bytes
		if memoryUsage > 0 {
			memoryUsage = memoryUsage / 1000 / 1000 // convert to mb
		}
		if memoryMaxUsage > 0 {
			memoryMaxUsage = memoryMaxUsage / 1000 / 1000 // convert to mb
		}

		jsonString, _ := json.Marshal(map[string]interface{}{
			"cpu_usage_percent":   percentCpuUsage,
			"memory_usage_mb":     memoryUsage,
			"max_memory_usage_mb": memoryMaxUsage,
		})

		// non blocking channel sending
		select {
		case *jsonStats <- string(jsonString):
			break
		default:
			break
		}
	}
	return nil
}

func Close() {
	cli.Close()
}
