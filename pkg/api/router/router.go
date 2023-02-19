package router

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"net/http"
)

const Port = 80

func Register() *mux.Router {
	r := mux.NewRouter()
	r.Use(authMiddleware)
	r.HandleFunc("/", rootRoute).Methods("GET")

	r.HandleFunc("/login", loginRoute).Methods("POST")
	r.HandleFunc("/user/password/change", passwordChange).Methods("POST")

	r.HandleFunc("/server/prepared", getPreparedServer).Methods("GET")
	r.HandleFunc("/server/start", startServer).Methods("POST")
	r.HandleFunc("/server/start/status/{serverid}", serverStartStatus).Methods("GET")
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
