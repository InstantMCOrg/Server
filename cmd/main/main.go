package main

import (
	"fmt"
	"github.com/instantminecraft/server/pkg/config"
	"github.com/instantminecraft/server/pkg/manager"
)

func main() {
	manager.InitDockerSystem()
	defer manager.Close()
	container, err := manager.ListContainersByNameStart(config.CONTAINER_BASE_NAME)
	if err != nil {
		panic(err)
	}
	//manager.RunContainer(config.IMAGE_NAME, config.CONTAINER_BASE_NAME+"testname", 25555)
	fmt.Println(container)

}
