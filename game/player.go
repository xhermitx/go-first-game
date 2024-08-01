package game

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Player struct {
	PlayerId uuid.UUID       `json:"player_id"`
	Position int             `json:"position"`
	Conn     *websocket.Conn `json:"-"`
}

func NewPlayer(conn *websocket.Conn) *Player {
	return &Player{
		PlayerId: uuid.New(),
		Position: 0,
		Conn:     conn,
	}
}

func (p *Player) UpdatePosition(position int) {
	p.Position = position
}
