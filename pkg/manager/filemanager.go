package manager

import (
	"github.com/instantminecraft/server/pkg/config"
	"github.com/rs/zerolog/log"
	"os"
	"path/filepath"
)

// EnsureDirsExist Checks if all needed directories exist. If not they will be created
func EnsureDirsExist() {
	if _, err := os.Stat(config.DataDir); os.IsNotExist(err) {
		// Create missing directory
		if err := os.Mkdir(config.DataDir, os.ModePerm); err != nil {
			log.Fatal().Err(err).Msgf("Couldn't create the directory %s", config.DataDir)
		}
	}
}

func DeleteMcWorld(port int) error {
	return os.RemoveAll(filepath.Join(config.DataDir, string(port)))
}
