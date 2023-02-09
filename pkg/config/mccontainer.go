package config

const (
	BaseImageName     = "ghcr.io/instantminecraft/client"
	ContainerBaseName = "MC-Server-"
	McServerProxyPort = 25585
)

var (
	AvailableVersions         = [...]string{"1.19.3", "1.19.2", "1.19", "1.18.2", "1.18", "1.17", "1.16.5", "1.16", "1.15.2", "1.15", "1.14.4", "1.14", "1.13.2", "1.13", "1.12.2", "1.12", "1.11.2", "1.11", "1.10.2", "1.9.4", "1.9", "1.8.9", "1.8", "1.7.10", "1.7"}
	WaitingReadyContainerName = ContainerBaseName + "ready"
	LatestImageName           = "ghcr.io/instantminecraft/client:mc-" + AvailableVersions[0]
)
