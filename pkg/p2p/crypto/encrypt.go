package crypto

import (
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
)

// Encryptor handles message encryption and decryption
type Encryptor struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewEncryptor creates a new encryptor with generated keys
func NewEncryptor() (*Encryptor, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	return &Encryptor{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}, nil
}

// GenerateKeyPair generates a new RSA key pair
func GenerateKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	return privateKey, &privateKey.PublicKey, nil
}

// EncryptMessage encrypts a message using AES-GCM with RSA-encrypted key
func (e *Encryptor) EncryptMessage(plaintext []byte, recipientPubKey *rsa.PublicKey) ([]byte, error) {
	// Generate a random AES key
	aesKey := make([]byte, 32) // 256-bit key
	if _, err := io.ReadFull(rand.Reader, aesKey); err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}

	// Encrypt the AES key with the recipient's public key
	encryptedAESKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, recipientPubKey, aesKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt AES key: %w", err)
	}

	// Encrypt the plaintext with AES-GCM
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Combine encrypted key + ciphertext
	result := make([]byte, len(encryptedAESKey)+len(ciphertext))
	copy(result[:len(encryptedAESKey)], encryptedAESKey)
	copy(result[len(encryptedAESKey):], ciphertext)

	return result, nil
}

// DecryptMessage decrypts a message using RSA to decrypt AES key and AES-GCM to decrypt message
func (e *Encryptor) DecryptMessage(encryptedData []byte, senderPubKey *rsa.PublicKey) ([]byte, error) {
	keySize := e.privateKey.Size()
	if len(encryptedData) < keySize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	// Extract encrypted AES key
	encryptedAESKey := encryptedData[:keySize]
	ciphertext := encryptedData[keySize:]

	// Decrypt the AES key with our private key
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, e.privateKey, encryptedAESKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AES key: %w", err)
	}

	// Decrypt the ciphertext with AES-GCM
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt message: %w", err)
	}

	return plaintext, nil
}

// SignMessage signs a message with the private key
func (e *Encryptor) SignMessage(message []byte) ([]byte, error) {
	hash := sha256.Sum256(message)
	signature, err := rsa.SignPSS(rand.Reader, e.privateKey, crypto.SHA256, hash[:], nil)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}

	return signature, nil
}

// VerifySignature verifies a signature against a message and public key
func (e *Encryptor) VerifySignature(message, signature []byte, pubKey *rsa.PublicKey) error {
	hash := sha256.Sum256(message)
	err := rsa.VerifyPSS(pubKey, crypto.SHA256, hash[:], signature, nil)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// MarshalPublicKey converts a public key to PEM format
func MarshalPublicKey(pubKey *rsa.PublicKey) ([]byte, error) {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return pubKeyPEM, nil
}

// UnmarshalPublicKey converts PEM format to a public key
func UnmarshalPublicKey(pubKeyPEM []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pubKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPubKey, nil
}

// MarshalPrivateKey converts a private key to PEM format
func MarshalPrivateKey(privKey *rsa.PrivateKey) ([]byte, error) {
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKeyBytes,
	})

	return privKeyPEM, nil
}

// UnmarshalPrivateKey converts PEM format to a private key
func UnmarshalPrivateKey(privKeyPEM []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return privKey, nil
}