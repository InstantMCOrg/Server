package config

const (
	IMAGE_NAME          = "ghcr.io/instantminecraft/client:latest"
	BASE_IMAGE_NAME     = "ghcr.io/instantminecraft/client"
	CONTAINER_BASE_NAME = "MC-Server-"
	MCSERVER_PROXY_PORT = 25585
)

var (
	AVAILABLE_VERSIONS = [...]string{"1.19.3", "1.19.2", "1.19", "1.18.2", "1.18", "1.17", "1.16.5", "1.16", "1.15.2", "1.15", "1.14.4", "1.14", "1.13.2", "1.13", "1.12.2", "1.12", "1.11.2", "1.11", "1.10.2", "1.9.4", "1.9", "1.8.9", "1.8", "1.7.10", "1.7"}
)
