package router

import (
	"github.com/instantmc/server/pkg/db"
	"github.com/instantmc/server/pkg/models"
	"net/http"
)

func getCurrentUser(r *http.Request) (models.User, error) {
	clientAuthKey := r.Header.Get("auth")
	// searching for session...
	return db.GetUserFromToken(clientAuthKey)
}
