package router

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"net/http"
)

const Port = 25000

func Register() *mux.Router {
	r := mux.NewRouter()
	api := r.PathPrefix("/api").Subrouter()
	api.Use(authMiddleware)
	api.HandleFunc("/", rootRoute).Methods("GET")

	api.HandleFunc("/login", loginRoute).Methods("POST")
	api.HandleFunc("/user/password/change", passwordChange).Methods("POST")

	api.HandleFunc("/server", getServer).Methods("GET")
	api.HandleFunc("/server/prepared", getPreparedServer).Methods("GET")
	api.HandleFunc("/server/start", startServer).Methods("POST")
	api.HandleFunc("/server/start/status/{serverid}", serverStartStatus).Methods("GET")
	api.HandleFunc("/server/stats/{serverid}", serverStats).Methods("GET")
	api.HandleFunc("/server/{serverid}/delete", deleteServer).Methods("DELETE")

	// Flutter frontend
	fs := http.FileServer(http.Dir("./frontend/"))
	r.PathPrefix("/").Handler(fs)

	return r
}

func HandleHttpRequests() {
	router := Register()
	for true { // Handle forever
		log.Info().Msgf("Starting Http Server on port %d...", Port)
		err := http.ListenAndServe(fmt.Sprintf(":%d", Port), router)
		if err != nil {
			log.Error().Err(err).Msg("An error occurred while serving the Http endpoint")
		}
	}
}
