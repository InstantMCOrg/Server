package models

type MCServer struct {
	ID     string
	Port   int
	UserID string
}

type ServerStatus struct {
	Server struct {
		Running bool `json:"running"`
	} `json:"server"`
}
