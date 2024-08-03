//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"github.com/xhermitx/go-first-game/game"
)

const (
	baseURL   = "http://localhost:8000/game/"
	socketURL = "ws://localhost:8000/ws/"
)

type Game struct {
	GameId       string      `json:"game_id"`
	Participants []Player    `json:"participants"`
	Text         string      `json:"text"`
	Status       game.Status `json:"status"`
	Winner       Player      `json:"winner"`
}

func (g *Game) AddPlayer(player Player) {
	g.Participants = append(g.Participants, player)
}

type Player struct {
	PlayerId string `mapstructure:"player_id"`
	Position int    `mapstructure:"position"`
}

func socketConn(gameId string) (*websocket.Conn, error) {
	url := socketURL + gameId
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return conn, err
}

var (
	players = make(map[string]Player)
	lock    = sync.Mutex{}
)

func main() {
	newGame, err := newGame()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Game ID:", newGame.GameId)

	conn, err := socketConn(newGame.GameId)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println(newGame.Status)

	go func() {
		for {
			var msg map[string]any
			if err := conn.ReadJSON(&msg); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				} else {
					log.Printf("ReadJSON error: %v", err)
				}
				return
			}
			log.Println("Message received: ", msg)

			lock.Lock()
			defer lock.Unlock()
			if err := handleMsg(conn, newGame, msg); err != nil {
				log.Fatal(err)
			}
		}
	}()

	time.Sleep(time.Second * 3)

	for newGame.Status == game.InProgress {
		time.Sleep(time.Millisecond * 500)
		curPosition := rand.Intn(len(newGame.Text) + 1)
		msg := game.NewMessage(game.PositionBroadcast, struct{ Position int }{curPosition})

		fmt.Println("Sending: ", msg.Type, msg.Payload)

		if err := conn.WriteJSON(msg); err != nil {
			log.Println("error writing the message: ", err)
		}
	}
}

func handleMsg(conn *websocket.Conn, g *Game, msg map[string]any) error {

	payload, ok := msg["payload"].(map[string]any)
	if !ok {
		return fmt.Errorf("payload is not a map[string]any")
	}

	switch game.MessageType(msg["type"].(string)) {

	case game.AddPlayer:
		var player Player
		if err := mapstructure.Decode(payload, &player); err != nil {
			return err
		}
		g.AddPlayer(player)
		fmt.Println("New player: ", player.PlayerId, " at position: ", player.Position)

		if len(g.Participants) == 2 {
			msg := game.NewMessage(game.MessageType(game.StartGame), g)
			if err := conn.WriteJSON(msg); err != nil {
				return err
			}
		}

	case game.PositionBroadcast:
		var player Player
		if err := mapstructure.Decode(payload, &player); err != nil {
			return err
		}
		players[player.PlayerId] = player
		fmt.Println("Opponent Position: ", player.Position)

	case game.StatusUpdate:
		if err := mapstructure.Decode(payload, g); err != nil {
			return err
		}
		if g.Status == game.TheEnd {
			fmt.Println("Winner is: ", g.Winner)
		}
	default:
		for k, v := range msg {
			log.Println(k, v)
		}
		return fmt.Errorf("Unknown message type: %s", msg["type"])
	}

	return nil
}

func newGame() (*Game, error) {
	resp, err := http.Get(baseURL + "create")
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var game Game
	if err := json.Unmarshal(body, &game); err != nil {
		return nil, err
	}

	return &game, nil
}

func joinGame(gameId string) (*Game, error) {

	data := map[string]string{"game_id": gameId}
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(baseURL+"join", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var game Game
	if err = json.NewDecoder(resp.Body).Decode(&game); err != nil {
		return nil, err
	}

	return &game, nil
}
