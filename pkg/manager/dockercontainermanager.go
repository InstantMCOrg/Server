package manager

import (
	"bufio"
	"strconv"
	"strings"
	"time"

	"github.com/docker/distribution/context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/instantminecraft/server/pkg/config"
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
func RunContainer(imageName string, containerName string, containerPort int, env []string) (string, error) {
	port := strconv.Itoa(config.McServerProxyPort) + "/tcp"
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

func IsContainerPaused(containerID string) (bool, error) {
	containerStats, err := GetContainerStats(containerID)
	if err != nil {
		return false, err
	}
	return containerStats.State.Paused, nil
}

func KillContainer(containerID string) error {
	return cli.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{Force: true})
}

func Close() {
	cli.Close()
}
