package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
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
		return
	}
	defer r.Body.Close()

	var temp struct {
		GameId uuid.UUID `json:"game_id"`
	}

	if err = json.Unmarshal(body, &temp); err != nil {
		// handle error
		return
	}

	if resGame, ok := h.Arenas[temp.GameId]; !ok {
		http.Error(w, "game not found", http.StatusNotFound)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resGame)
	}
}

func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {

	h.Lock()
	defer h.Unlock()

	vars := mux.Vars(r)
	payload := vars["game_id"]

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

	if err := activeGame.BroadCast(game.NewMessage(game.AddPlayer, newPlayer)); err != nil {
		// handle error
		_ = err
	}

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

	case game.MessageType(game.StartGame):
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
