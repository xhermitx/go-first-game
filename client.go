//go:build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/xhermitx/go-first-game/game"
)

const (
	baseURL   = "http://localhost:8000"
	socketURL = "ws://localhost:8000"
)

type Game struct {
	GameId       string      `json:"game_id"`
	Participants []Player    `json:"participants"`
	Text         string      `json:"text"`
	Status       game.Status `json:"status"`
}

type Player struct {
	PlayerId string `json:"player_id"`
	Position int    `json:"position"`
}

func socketConn(gameId string) (*websocket.Conn, error) {
	url := socketURL + gameId
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return conn, err
}

func main() {
	newGame, err := newGame()
	if err != nil {
		log.Fatal(err)
	}

	conn, err := socketConn(newGame.GameId)
	if err != nil {
		log.Panic(err)
	}
	defer conn.Close()

	go func() {
		for {
			var msg game.Message
			if err := conn.ReadJSON(&msg); err != nil {
				log.Println(err)
				return
			}
			log.Println(msg)

			// handle msg
		}
	}()

	for {
		var input int
		fmt.Println("Enter message")
		fmt.Scan(&input)

		msg := game.NewMessage(game.MessageType(game.StartGame), nil)
		err := conn.WriteJSON()
	}
}

func handleMsg(g *Game, msg game.Message) error {

	switch msg.Type {

	case 

	case game.MessageType(game.AddPlayer):
		player := msg.Payload.(Player)
		g.AddPlayer(player)
		fmt.Println("New player: ", player.PlayerId, " at position: ", player.Position)

	case game.MessageType(game.PositionBroadcast):

	}
}

func newGame() (*Game, error) {
	resp, err := http.Get(baseURL + "/game/create/")
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var game *Game

	if err := json.Unmarshal(body, &game); err != nil {
		return nil, err
	}

	return game, nil
}

func joinGame(gameId string) (*Game, error) {

	data := map[string]string{"game_id": gameId}
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(baseURL+"/join/", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var game *Game
	if err = json.NewDecoder(resp.Body).Decode(game); err != nil {
		return nil, err
	}

	return game, nil
}
