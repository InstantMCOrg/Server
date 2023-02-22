package router

import (
	"github.com/instantmc/server/pkg/db"
	"net/http"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path != "/" && r.URL.Path != "/login" {
			clientAuthKey := r.Header.Get("auth")
			// searching for session...
			_, err := db.GetSession(clientAuthKey)

			if err != nil {
				// Authentication failed
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}
