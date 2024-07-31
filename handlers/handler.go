package handlers

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/xhermitx/go-first-game/game"
)

type Handler struct {
	Arenas map[uuid.UUID]*game.Game
	sync.Mutex
}

func NewHandler() *Handler {
	return &Handler{
		Arenas: make(map[uuid.UUID]*game.Game),
	}
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

	h.Lock()
	defer h.Unlock()

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
	activeGame := h.Arenas[gameId]

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// handle error
		return
	}
	defer conn.Close()

	newPlayer := game.NewPlayer(conn)
	activeGame.AddPlayer(*newPlayer)

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			// handle err
			_ = err
		}

		var msg game.Message
		if err := json.Unmarshal(data, &msg); err != nil {
			// handle err
			_ = err
		}

		if err := h.HandleGame(activeGame, newPlayer, msg); err != nil {
			// handle error
			return
		}
	}
}

func (h *Handler) HandleGame(g *game.Game, player *game.Player, msg game.Message) error {

	switch msg.Type {
	case game.MessageType(game.Created):
		g.UpdateStatus(nil)

	case game.MessageType(game.PositionBroadcast):

		position := msg.Payload.(struct{ Position int }).Position

		player.UpdatePosition(position)

		if position == len(g.Text) {
			// Check for Winner and broadcast
			g.UpdateStatus(player)
			delete(h.Arenas, g.GameId) // Remove the game from the list
		}

		posUpdate := game.Message{
			Type:    msg.Type,
			Payload: player,
		}
		if err := g.BroadCast(posUpdate); err != nil {
			return err
		}
	}
	return nil
}
