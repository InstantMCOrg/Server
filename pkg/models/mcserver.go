package models

import (
	"github.com/instantminecraft/server/pkg/enums"
	"sync"
)

type McServerContainer struct {
	ServerID    string             `json:"server_id"`
	Name        string             `json:"name"`
	ContainerID string             `json:"container_id"`
	McVersion   string             `json:"mc_version"`
	RamSizeMB   int                `json:"ram_size_mb"`
	Port        int                `json:"port"`
	Status      enums.ServerStatus `json:"Status"`
}

func (mcServer *McServerContainer) Self() *McServerContainer {
	return mcServer
}

func (mcServer *McServerContainer) ToClientJson() interface{} {
	return struct {
		ServerID  string `json:"server_id"`
		Name      string `json:"name"`
		McVersion string `json:"mc_version"`
		Port      int    `json:"port"`
		RamSizeMB int    `json:"ram_size_mb"`
		Status    string `json:"status"`
	}{
		ServerID:  mcServer.ServerID,
		Name:      mcServer.Name,
		McVersion: mcServer.McVersion,
		Port:      mcServer.Port,
		RamSizeMB: mcServer.RamSizeMB,
		Status:    mcServer.Status.String(),
	}
}

type ServerStatus struct {
	Server struct {
		Running bool `json:"running"`
	} `json:"server"`
}

// McServerPreparationConfig
// ´CoreBootUpWG´ waits until the http server started
type McServerPreparationConfig struct {
	Port         int
	AuthKey      string
	RamSizeMB    int
	CoreBootUpWG *sync.WaitGroup
}

// McContainerSearchConfig
// Default Status is Prepared
type McContainerSearchConfig struct {
	McVersion string
	Status    enums.ServerStatus
	RamSizeMB int
}
