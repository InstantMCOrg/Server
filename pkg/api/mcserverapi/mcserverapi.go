package mcserverapi

import (
	"encoding/json"
	"fmt"
	"github.com/instantminecraft/server/pkg/models"
	"net/http"
)

const authHeader = "auth"

func GetServerStatus(port int, authKey string) (models.ServerStatus, error) {
	client := &http.Client{}
	url := fmt.Sprintf("http://localhost:%d", port)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set(authHeader, authKey)
	resp, err := client.Do(req)

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

func WaitForMcWorldBootUp(port int, authKey string) error {
	client := &http.Client{}
	url := fmt.Sprintf("http://localhost:%d/server/start?blocking=true", port)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set(authHeader, authKey)
	_, err := client.Do(req)

	if err != nil {
		return err
	}
	return nil
}
