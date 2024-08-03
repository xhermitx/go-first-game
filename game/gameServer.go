package game

import (
	"github.com/google/uuid"
)

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
	Winner       Player    `json:"winner"`
}

func NewGame() *Game {
	return &Game{
		GameId:       uuid.New(),
		Participants: nil,
		Status:       Created,
		Text:         genRandomText(),
	}
}

func (g *Game) BroadCast(msgType MessageType, payload any) error {

	msg := NewMessage(msgType, payload)

	for _, player := range g.Participants {
		if err := player.Conn.WriteJSON(msg); err != nil {
			return err
		}
	}
	return nil
}

func (g *Game) AddPlayer(player Player) {
	g.Participants = append(g.Participants, player)
}

func (g *Game) UpdateStatus(winner *Player) error {

	switch g.Status {

	case Created:
		g.Status = InProgress
		if err := g.BroadCast(StatusUpdate, *g); err != nil {
			return err
		}

	case InProgress:
		g.Status = TheEnd
		if err := g.BroadCast(StatusUpdate, *g); err != nil {
			return err
		}
	}

	return nil
}

func genRandomText() string {
	// To be implemented
	return "This is a sample Text"
}
