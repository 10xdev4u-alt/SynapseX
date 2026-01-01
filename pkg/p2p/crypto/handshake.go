package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

// HandshakeMessage represents a message used in the secure handshake
type HandshakeMessage struct {
	NodeID      string `json:"node_id"`
	PublicKey   []byte `json:"public_key"`
	Timestamp   int64  `json:"timestamp"`
	Signature   []byte `json:"signature"`
	SessionKey  []byte `json:"session_key,omitempty"`
}

// HandshakeManager handles secure handshake protocol
type HandshakeManager struct {
	encryptor *Encryptor
	nodeID    string
}

// NewHandshakeManager creates a new handshake manager
func NewHandshakeManager(encryptor *Encryptor, nodeID string) *HandshakeManager {
	return &HandshakeManager{
		encryptor: encryptor,
		nodeID:    nodeID,
	}
}

// CreateHandshakeMessage creates a signed handshake message
func (h *HandshakeManager) CreateHandshakeMessage() (*HandshakeMessage, error) {
	pubKeyPEM, err := MarshalPublicKey(h.encryptor.publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	// Create a random session key for this session
	sessionKey := make([]byte, 32)
	if _, err := rand.Read(sessionKey); err != nil {
		return nil, fmt.Errorf("failed to generate session key: %w", err)
	}

	msg := &HandshakeMessage{
		NodeID:     h.nodeID,
		PublicKey:  pubKeyPEM,
		Timestamp:  time.Now().Unix(),
		SessionKey: sessionKey,
	}

	// Sign the message
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	signature, err := h.encryptor.SignMessage(msgBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}

	msg.Signature = signature
	return msg, nil
}

// VerifyHandshakeMessage verifies a received handshake message
func (h *HandshakeManager) VerifyHandshakeMessage(msg *HandshakeMessage) error {
	if msg == nil {
		return fmt.Errorf("handshake message is nil")
	}

	// Unmarshal the public key
	pubKey, err := UnmarshalPublicKey(msg.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	// Create a copy of the message without the signature for verification
	msgCopy := &HandshakeMessage{
		NodeID:     msg.NodeID,
		PublicKey:  msg.PublicKey,
		Timestamp:  msg.Timestamp,
		SessionKey: msg.SessionKey,
	}

	// Marshal the message copy
	msgBytes, err := json.Marshal(msgCopy)
	if err != nil {
		return fmt.Errorf("failed to marshal message copy: %w", err)
	}

	// Verify the signature
	if err := h.encryptor.VerifySignature(msgBytes, msg.Signature, pubKey); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	// Check timestamp (within 5 minutes)
	currentTime := time.Now().Unix()
	if currentTime-msg.Timestamp > 300 || msg.Timestamp-currentTime > 300 {
		return fmt.Errorf("timestamp is too old or too far in the future")
	}

	return nil
}

// EncryptHandshakeMessage encrypts a handshake message
func (h *HandshakeManager) EncryptHandshakeMessage(msg *HandshakeMessage, recipientPubKey *rsa.PublicKey) ([]byte, error) {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal handshake message: %w", err)
	}

	return h.encryptor.EncryptMessage(msgBytes, recipientPubKey)
}

// DecryptHandshakeMessage decrypts a handshake message
func (h *HandshakeManager) DecryptHandshakeMessage(encryptedData []byte, senderPubKey *rsa.PublicKey) (*HandshakeMessage, error) {
	decryptedBytes, err := h.encryptor.DecryptMessage(encryptedData, senderPubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt handshake message: %w", err)
	}

	var msg HandshakeMessage
	if err := json.Unmarshal(decryptedBytes, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal handshake message: %w", err)
	}

	return &msg, nil
}

// CreateChallenge creates a challenge for authentication
func (h *HandshakeManager) CreateChallenge() ([]byte, error) {
	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		return nil, fmt.Errorf("failed to generate challenge: %w", err)
	}

	return challenge, nil
}

// SignChallenge signs a challenge with the private key
func (h *HandshakeManager) SignChallenge(challenge []byte) ([]byte, error) {
	hash := sha256.Sum256(challenge)
	signature, err := rsa.SignPSS(rand.Reader, h.encryptor.privateKey, crypto.SHA256, hash[:], nil)
	if err != nil {
		return nil, fmt.Errorf("failed to sign challenge: %w", err)
	}

	return signature, nil
}

// VerifyChallenge verifies a signed challenge
func (h *HandshakeManager) VerifyChallenge(challenge, signature []byte, pubKey *rsa.PublicKey) error {
	hash := sha256.Sum256(challenge)
	err := rsa.VerifyPSS(pubKey, crypto.SHA256, hash[:], signature, nil)
	if err != nil {
		return fmt.Errorf("challenge verification failed: %w", err)
	}

	return nil
}