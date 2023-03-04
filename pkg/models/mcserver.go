package models

import (
	"github.com/instantmc/server/pkg/enums"
	"sync"
	"time"
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
// CoreBootUpWG waits until the http server started
// If AutoDeploy is set to false the container will pause and wait until it is picked up
type McServerPreparationConfig struct {
	Port         int
	AuthKey      string
	RamSizeMB    int
	CoreBootUpWG *sync.WaitGroup
	ServerID     string
	AutoDeploy   bool
}

// McContainerSearchConfig
// Default Status is Prepared
// IF Status is enums.Running ready prepared container are NOT returned
type McContainerSearchConfig struct {
	McVersion string
	Status    enums.ServerStatus
	RamSizeMB int
}

type McContainerResourceStats struct {
	CpuUsage    float64   `json:"cpu_usage_percent"`
	MemoryUsage uint64    `json:"memory_usage_mb"`
	Time        time.Time `json:"time"`
}
