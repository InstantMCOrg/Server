package main

import (
	"github.com/instantmc/server/pkg/api/router"
	"github.com/instantmc/server/pkg/db"
	"github.com/instantmc/server/pkg/manager"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"
)

func main() {
	// Setup logger
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC822})

	// Ensures all needed directories exist
	manager.EnsureDirsExist()
	// Setup Database
	db.Init()
	manager.InitDockerSystem()
	defer manager.Close()
	manager.InitMCServerManagement()
	router.HandleHttpRequests()
}
