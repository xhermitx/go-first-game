package game

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type MessageType string

const (
	JoinGame         MessageType = "join_game"
	PositionBrodcast MessageType = "pos_broadcast"
	StatusUpdate     MessageType = "status_update"
)

type Message struct {
	Type    MessageType `json:"type"`
	Payload []byte      `json:"payload"`
}

func NewMessage(msgType MessageType, data any) (*Message, error) {

	payload, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &Message{
		Type:    msgType,
		Payload: payload,
	}, nil
}

type Status string

const (
	Created    Status = "Created"
	InProgress Status = "In Progress"
	TheEnd     Status = "The End"
)

type Game struct {
	GameId       uuid.UUID `json:"game_id"`
	Participants []Player  `json:"participants"`
	Text         string    `json:"text"`
	Status       Status    `json:"status"`
}

func NewGame() *Game {
	return &Game{
		GameId:       uuid.New(),
		Participants: nil,
		Status:       Created,
	}
}

func (g *Game) BroadCast(msg Message) error {
	for _, player := range g.Participants {
		if err := player.Conn.WriteJSON(msg); err != nil {
			return err
		}
	}
	return nil
}

func (g *Game) AddPlayer(player Player) error {
	g.Participants = append(g.Participants, player)
}

func (g *Game) UpdateGame() {
	// to be implemented
}

type Player struct {
	PlayerId uuid.UUID `json:"player_id"`
	Position int       `json:"position"`
	Conn     *websocket.Conn
}

func NewPlayer(conn *websocket.Conn) *Player {
	return &Player{
		PlayerId: uuid.New(),
		Position: 0,
		Conn:     conn,
	}
}
