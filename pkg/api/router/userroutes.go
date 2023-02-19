package router

import (
	"encoding/json"
	"github.com/instantminecraft/server/pkg/config"
	"github.com/instantminecraft/server/pkg/db"
	"github.com/instantminecraft/server/pkg/utils"
	"net/http"
)

func loginRoute(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		sendError("Please provide \"username\" and \"password\"", w, http.StatusBadRequest)
		return
	}

	// check if user exists
	hashedPassword := utils.SHA256([]byte(password))

	user, err := db.Login(username, hashedPassword)
	if err != nil {
		// User doesn't exist
		sendError("Invalid credentials", w, http.StatusUnauthorized)
		return
	}

	// creating session and token
	token, err := db.CreateSession(&user)
	if err != nil {
		// Couldn't create session
		sendError("Couldn't create session", w, http.StatusInternalServerError)
		return
	}

	// Session successfully created
	data, _ := json.Marshal(map[string]interface{}{
		"token":                    token,
		"password_change_required": password == config.PasswordRequiresChange,
	})
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
