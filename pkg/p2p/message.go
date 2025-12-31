package p2p

import (
	"encoding/json"
	"fmt"
	"time"
)

const (
	MessageTypeHello     = "HELLO"
	MessageTypePeerList  = "PEER_LIST"
	MessageTypeDataSync  = "DATA_SYNC"
	MessageTypeHeartbeat = "HEARTBEAT"
	MessageTypeError     = "ERROR"
)

// Message represents a P2P network message
type Message struct {
	Type      string      `json:"type"`
	ID        string      `json:"id"`
	Sender    string      `json:"sender"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload"`
}

// HelloPayload contains data for HELLO messages
type HelloPayload struct {
	NodeID      string `json:"node_id"`
	Version     string `json:"version"`
	ListenPort  int    `json:"listen_port"`
	Capabilities []string `json:"capabilities"`
}

// PeerListPayload contains data for PEER_LIST messages
type PeerListPayload struct {
	Peers []PeerInfo `json:"peers"`
}

// PeerInfo represents information about a peer
type PeerInfo struct {
	ID       string `json:"id"`
	Address  string `json:"address"`
	Version  string `json:"version"`
	LastSeen int64  `json:"last_seen"`
}

// DataSyncPayload contains data for DATA_SYNC messages
type DataSyncPayload struct {
	DataID    string      `json:"data_id"`
	Type      string      `json:"type"`
	Content   interface{} `json:"content"`
	Version   int64       `json:"version"`
	Timestamp int64       `json:"timestamp"`
}

// HeartbeatPayload contains data for HEARTBEAT messages
type HeartbeatPayload struct {
	NodeID string `json:"node_id"`
	TS     int64  `json:"timestamp"`
}

// ErrorPayload contains data for ERROR messages
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewMessage creates a new message with the given type and payload
func NewMessage(msgType string, sender string, payload interface{}) Message {
	return Message{
		Type:      msgType,
		ID:        fmt.Sprintf("%s-%d", msgType, time.Now().UnixNano()),
		Sender:    sender,
		Timestamp: time.Now(),
		Payload:   payload,
	}
}

// Serialize converts a message to JSON bytes
func (m *Message) Serialize() ([]byte, error) {
	return json.Marshal(m)
}

// DeserializeMessage converts JSON bytes to a message
func DeserializeMessage(data []byte) (*Message, error) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}
	return &msg, nil
}

// Validate checks if a message is valid
func (m *Message) Validate() error {
	if m.Type == "" {
		return fmt.Errorf("message type cannot be empty")
	}
	if m.ID == "" {
		return fmt.Errorf("message ID cannot be empty")
	}
	if m.Sender == "" {
		return fmt.Errorf("message sender cannot be empty")
	}
	return nil
}
