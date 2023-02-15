package router

import (
	"net/http"
)

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.URL.Path != "/" && r.URL.Path != "/login" {
			// Todo
			if r.Header.Get("auth") != "" {
				// Authentication failed
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
		}

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}
