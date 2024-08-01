package game

type MessageType string

const (
	StartGame         MessageType = "start_game"
	AddPlayer         MessageType = "add_player"
	PositionBroadcast MessageType = "pos_broadcast"
	StatusUpdate      MessageType = "status_update"
)

type Message struct {
	Type    MessageType `json:"type"`
	Payload any         `json:"payload"`
}

func NewMessage(msgType MessageType, data any) Message {
	return Message{
		Type:    msgType,
		Payload: data,
	}
}
