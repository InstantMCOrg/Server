package models

type McServerContainer struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ContainerID string `json:"container_id"`
	McVersion   string `json:"mc_version"`
	Port        int    `json:"port"`
	Running     bool   `json:"running"`
}

type ServerStatus struct {
	Server struct {
		Running bool `json:"running"`
	} `json:"server"`
}
