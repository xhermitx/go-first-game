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

func (g *Game) AddPlayer(player Player) {
	g.Participants = append(g.Participants, player)
}

func (g *Game) UpdateStatus(winner *Player) error {

	switch g.Status {
	case Created:
		g.Status = InProgress
		g.Text = genRandomText()

		msg := NewMessage(StatusUpdate, g)

		if err := g.BroadCast(msg); err != nil {
			return err
		}

	case InProgress:
		g.Status = TheEnd
		msg := NewMessage(StatusUpdate, winner)

		if err := g.BroadCast(msg); err != nil {
			return err
		}
	}

	return nil
}

func genRandomText() string {
	// To be implemented
	return "This is a sample Text"
}
