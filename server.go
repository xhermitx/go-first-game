package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	router.StrictSlash(true)

	h := NewHandler()

	router.HandleFunc("/game/create/", h.CreateGame).Methods("POST")
	router.HandleFunc("/game/join/", h.JoinGame).Methods("POST")
	router.HandleFunc("/ws/{game_id}", h.Connect)

	log.Panic(http.ListenAndServe(":8000", router))
}
