package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/xhermitx/go-first-game/game"
)

type Handler struct {
	Arenas map[uuid.UUID]*game.Game
}

func NewHandler() *Handler {
	return &Handler{}
}

var upgrader = websocket.Upgrader{}

func (h *Handler) CreateGame(w http.ResponseWriter, r *http.Request) {

	newGame := game.NewGame()

	h.Arenas[newGame.GameId] = newGame

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newGame)
}

func (h *Handler) JoinGame(w http.ResponseWriter, r *http.Request) {

	body, err := io.ReadAll(r.Body)
	if err != nil {
		// handle error
		log.Println(err)
		return
	}
	defer r.Body.Close()

	var data struct {
		GameId uuid.UUID `json:"game_id"`
	}

	if err = json.Unmarshal(body, &data); err != nil {
		// handle error
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(h.Arenas[data.GameId])
}

func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {

	payload := r.URL.Query().Get("game")
	if payload == "" {
		http.Error(w, "Missing game name", http.StatusBadRequest)
		return
	}

	gameId, err := uuid.Parse(payload)
	if err != nil {
		// handle error
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// handle error
		return
	}
	defer conn.Close()

	newPlayer := game.NewPlayer(conn)
	h.Arenas[gameId].AddPlayer(*newPlayer)

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			// handle err
			log.Println(err)
		}

		var msg game.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			// handle err
		}

		if err := game.HandleGame(msg); err != nil {
			// handle error
			return
		}
	}
}
