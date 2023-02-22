package manager

import (
	"github.com/instantmc/server/pkg/config"
	"github.com/instantmc/server/pkg/utils"
)

var usedPorts []int

func GeneratePort() int {
	port := utils.CreateRandomIntRange(config.PortRangeBegin, config.PortRangeEnd)
	if IsPortBeingUsed(port) {
		return GeneratePort()
	}
	return port
}

func IsPortBeingUsed(port int) bool {
	for _, usedPort := range usedPorts {
		if usedPort == port {
			return true
		}
	}
	return false
}

func AddPortToUsageList(port int) {
	usedPorts = append(usedPorts, port)
}

func RemovePortFromUsageList(port int) {
	var tempUsedPorts []int

	for _, usedPort := range usedPorts {
		if usedPort != port {
			tempUsedPorts = append(tempUsedPorts, usedPort)
		}
	}
	usedPorts = tempUsedPorts
}
