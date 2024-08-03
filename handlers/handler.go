package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
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
		http.Error(w, "bad request: game_id", http.StatusBadRequest)
		return
	}
	activeGame := h.Arenas[gameId]

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// handle error
		http.Error(w, "connection failed", http.StatusInternalServerError)
		return
	}

	newPlayer := game.NewPlayer(conn)
	activeGame.AddPlayer(newPlayer)

	if err := activeGame.BroadCast(game.AddPlayer, newPlayer); err != nil {
		// handle error
		http.Error(w, "error broadcasting", http.StatusInternalServerError)
		return
	}

	for {
		var data map[string]any
		if err := conn.ReadJSON(&data); err != nil {
			// handle err
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			} else {
				log.Printf("ReadJSON error: %v", err)
			}
			break
		}

		if err := h.HandleGame(activeGame, newPlayer, data); err != nil {
			// handle error
			log.Fatal("error handling the game: ", err)
			return
		}
	}

}

func (h *Handler) HandleGame(g *game.Game, player game.Player, msg map[string]any) error {

	payload, ok := msg["payload"].(map[string]any)
	if !ok {
		return fmt.Errorf("payload is not of type map[string]any")
	}

	switch msg["type"] {

	case game.StartGame:
		if err := g.UpdateStatus(nil); err != nil {
			return err
		}

	case game.PositionBroadcast:

		if err := mapstructure.Decode(payload, &player); err != nil {
			return err
		}

		if player.Position == len(g.Text) {
			// Check for Winner and broadcast
			if err := g.UpdateStatus(&player); err != nil {
				return err
			}
			delete(h.Arenas, g.GameId) // Remove the game from the list
		}

		if err := g.BroadCast(game.PositionBroadcast, player); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unknown message type: %s", msg["type"])
	}
	return nil
}
