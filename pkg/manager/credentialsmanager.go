package manager

import "github.com/instantminecraft/server/pkg/utils"

var authKeyMap = map[string]string{}

const authKeyLength = 128

// GenerateAuthKeyForMcServer Generates an auth key for the mc docker container and saves it in memory
func GenerateAuthKeyForMcServer() string {
	key := utils.RandomString(authKeyLength)
	return key
}

// SaveAuthKey Saves the authKey combined with containerID to memory
func SaveAuthKey(containerID string, authKey string) {
	authKeyMap[containerID] = authKey
}

// GetAuthKeyForMcServer Returns the auth key for a server. If no auth key found an empty string is returned
// You need to save the auth key with `SaveAuthKey` before accessing this function
func GetAuthKeyForMcServer(containerID string) string {
	return authKeyMap[containerID]
}
