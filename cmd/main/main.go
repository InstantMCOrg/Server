package main

import (
	"github.com/instantminecraft/server/pkg/api/router"
	"github.com/instantminecraft/server/pkg/db"
	"github.com/instantminecraft/server/pkg/manager"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

func main() {
	// Setup logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC822})

	// Setup Database
	db.Init()
	manager.InitDockerSystem()
	defer manager.Close()
	manager.InitMCServerManagement()
	router.HandleHttpRequests()
}
