package utils

import (
	"github.com/docker/docker/api/types"
	"github.com/instantminecraft/server/pkg/config"
	"strings"
)

func GetMcVersionFromContainer(container types.Container) string {
	splitted := strings.Split(container.Image, config.BaseVanillaMcImageName)
	return splitted[len(splitted)-1]
}

func GetPortFromContainer(container types.Container) int {
	return int(container.Ports[0].PublicPort)
}
