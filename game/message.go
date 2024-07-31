package game

import "encoding/json"

type MessageType string

const (
	StartGame         MessageType = "start_game"
	PositionBroadcast MessageType = "pos_broadcast"
	StatusUpdate      MessageType = "status_update"
)

type Message struct {
	Type    MessageType `json:"type"`
	Payload any         `json:"payload"`
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
