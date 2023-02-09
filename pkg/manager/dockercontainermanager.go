package manager

import (
	"fmt"
	"github.com/docker/distribution/context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/instantminecraft/server/pkg/config"
	"strconv"
	"strings"
	"time"
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
	_, err := cli.ImagePull(ctx, config.LatestImageName, types.ImagePullOptions{})
	if err != nil {
		fmt.Println("Oh oh! An error occurred while downloading the newest", config.LatestImageName, "image:", err, "\nRetrying in 2 Seconds...")
		time.Sleep(2 * time.Second)
		ensureMCServerImageIsReady()
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

func IsContainerPaused(containerID string) (bool, error) {
	containerStats, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return false, err
	}
	return containerStats.State.Paused, nil
}

func Close() {
	cli.Close()
}
