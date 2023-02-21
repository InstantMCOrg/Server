package manager

import (
	"github.com/instantminecraft/server/pkg/config"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
	"strconv"
)

// EnsureDirsExist Checks if all needed directories exist. If not they will be created
func EnsureDirsExist() {
	if _, err := os.Stat(config.DataDir); os.IsNotExist(err) {
		// Create missing directory
		if err := os.Mkdir(config.DataDir, os.ModePerm); err != nil {
			log.Fatal().Err(err).Msgf("Couldn't create the directory %s", config.DataDir)
		}
		if err := os.MkdirAll(filepath.Join(config.DataDir, config.McWorldsDir), os.ModePerm); err != nil {
			log.Fatal().Err(err).Msgf("Couldn't create the directory %s", filepath.Join(config.DataDir, config.McWorldsDir))
		}
	}
}

func DeleteMcWorld(port int) error {
	path := filepath.Join(config.DataDir, config.McWorldsDir, strconv.Itoa(port))
	log.Info().Msgf("Deleting mc world %s...", path)
	return os.RemoveAll(path)
}
