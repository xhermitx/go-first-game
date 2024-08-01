//go:build ignore

package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/xhermitx/go-first-game/handlers"
)

func main() {
	router := mux.NewRouter()

	router.StrictSlash(true)

	h := handlers.NewHandler()

	router.HandleFunc("/game/create/", h.CreateGame).Methods("GET")
	router.HandleFunc("/game/join/", h.JoinGame).Methods("POST")
	router.HandleFunc("/ws/{game_id}", h.Connect)

	log.Panic(http.ListenAndServe(":8000", router))
}
