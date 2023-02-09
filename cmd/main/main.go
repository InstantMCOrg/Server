package main

import (
	"fmt"
	"github.com/instantminecraft/server/pkg/manager"
	"time"
)

func main() {
	manager.InitDockerSystem()
	defer manager.Close()
	manager.InitMCServerManagement()
	fmt.Println("Waiting...")
	manager.WaitForFinsishedPreparing()
	fmt.Println("Done")
	time.Sleep(5 * time.Second)
	manager.StartMcServer()

	select {} // DEBUG
}
