package router

import (
	"github.com/instantminecraft/server/pkg/db"
	"github.com/instantminecraft/server/pkg/models"
	"net/http"
)

func getCurrentUser(r *http.Request) (models.User, error) {
	clientAuthKey := r.Header.Get("auth")
	// searching for session...
	return db.GetUserFromToken(clientAuthKey)
}
