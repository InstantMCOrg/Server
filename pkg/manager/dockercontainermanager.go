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
	_, err := cli.ImagePull(ctx, config.IMAGE_NAME, types.ImagePullOptions{})
	if err != nil {
		fmt.Println("Oh oh! An error occurred while downloading the newest", config.IMAGE_NAME, "image:", err, "\nRetrying in 2 Seconds...")
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

func RunContainer(imageName string, containerName string, containerPort int) {
	port := strconv.Itoa(config.MCSERVER_PROXY_PORT) + "/tcp"
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		ExposedPorts: nat.PortSet{
			nat.Port(port): {},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			nat.Port(port): []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: strconv.Itoa(containerPort)}},
		},
	}, nil, nil, containerName)
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	fmt.Println(resp.ID)
}

func Close() {
	cli.Close()
}
