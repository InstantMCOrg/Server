package models

import "github.com/instantminecraft/server/pkg/enums"

type McServerContainer struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	ContainerID string             `json:"container_id"`
	McVersion   string             `json:"mc_version"`
	Port        int                `json:"port"`
	Status      enums.ServerStatus `json:"Status"`
}

func (mcServer *McServerContainer) ToClientJson() interface{} {
	return struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		McVersion string `json:"mc_version"`
		Port      int    `json:"port"`
		Status    string `json:"status"`
	}{
		ID:        mcServer.ID,
		Name:      mcServer.Name,
		McVersion: mcServer.McVersion,
		Port:      mcServer.Port,
		Status:    mcServer.Status.String(),
	}
}

type ServerStatus struct {
	Server struct {
		Running bool `json:"running"`
	} `json:"server"`
}
