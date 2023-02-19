package models

import (
	"github.com/instantminecraft/server/pkg/enums"
	"sync"
)

type McServerContainer struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	ContainerID string             `json:"container_id"`
	McVersion   string             `json:"mc_version"`
	RamSizeMB   int                `json:"ram_size_mb"`
	Port        int                `json:"port"`
	Status      enums.ServerStatus `json:"Status"`
}

func (mcServer *McServerContainer) ToClientJson() interface{} {
	return struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		McVersion string `json:"mc_version"`
		Port      int    `json:"port"`
		RamSizeMB int    `json:"ram_size_mb"`
		Status    string `json:"status"`
	}{
		ID:        mcServer.ID,
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
