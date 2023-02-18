package models

type McServerContainer struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ContainerID string `json:"container_id"`
	McVersion   string `json:"mc_version"`
	Port        int    `json:"port"`
	Running     bool   `json:"running"`
}

func (mcServer *McServerContainer) ToClientJson() interface{} {
	return struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		McVersion string `json:"mc_version"`
		Port      int    `json:"port"`
		Running   bool   `json:"running"`
	}{
		ID:        mcServer.ID,
		Name:      mcServer.Name,
		McVersion: mcServer.McVersion,
		Port:      mcServer.Port,
		Running:   mcServer.Running,
	}
}

type ServerStatus struct {
	Server struct {
		Running bool `json:"running"`
	} `json:"server"`
}
