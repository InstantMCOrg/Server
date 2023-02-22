package mcserverapi

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/instantmc/server/pkg/models"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

func GetWorldGenerationChan(port int, authKey string) (chan int, error) {
	host := fmt.Sprintf("localhost:%d", port)
	u := url.URL{Scheme: "ws", Host: host, Path: "/server/world/creation_status"}

	header := http.Header{}
	header.Add("auth", authKey)
	connection, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return nil, err
	}

	worldGenerationChan := make(chan int)

	go func() {
		for {
			var data map[string]interface{}
			connection.ReadJSON(&data)
			if data["status"] == "already running" {
				worldGenerationChan <- 100
				break
			} else if data["status"] == "preparing" {
				rawStatusString := fmt.Sprintf("%v", data["world_status"])
				status, _ := strconv.Atoi(rawStatusString)
				select {
				case worldGenerationChan <- status:
					break
				default:
					break
				}
				if status == 100 {
					break
				}
			}
		}
		connection.Close()
	}()

	return worldGenerationChan, nil
}

func SendMessage(port int, authKey string, message string) error {
	client := &http.Client{}
	targetUrl := fmt.Sprintf("http://localhost:%d/server/message/send", port)

	form := url.Values{}
	form.Add("message", message)

	req, _ := http.NewRequest("POST", targetUrl, strings.NewReader(form.Encode()))
	req.Header.Set(authHeader, authKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	_, err := client.Do(req)

	if err != nil {
		return err
	}
	return nil
}
