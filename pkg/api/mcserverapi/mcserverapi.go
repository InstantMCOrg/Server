package mcserverapi

import (
	"encoding/json"
	"github.com/instantminecraft/server/pkg/models"
	"net/http"
	"strconv"
)

func GetServerStatus(port int) (models.ServerStatus, error) {
	resp, err := http.Get("http://localhost:" + strconv.Itoa(port))
	if err != nil {
		return models.ServerStatus{}, err
	}
	defer resp.Body.Close()
	var serverResponse models.ServerStatus
	err = json.NewDecoder(resp.Body).Decode(&serverResponse)
	if err != nil {
		return models.ServerStatus{}, err
	}

	return serverResponse, err
}

func WaitForMcWorldBootUp(port int) error {
	_, err := http.Get("http://localhost:" + strconv.Itoa(port) + "/server/start?blocking=true")
	if err != nil {
		return err
	}
	return nil
}
