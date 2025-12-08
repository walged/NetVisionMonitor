package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var masterKey []byte

// Initialize loads or generates the master encryption key
func Initialize(dataDir string) error {
	keyPath := filepath.Join(dataDir, ".key")

	// Try to load existing key
	if data, err := os.ReadFile(keyPath); err == nil {
		key, err := hex.DecodeString(string(data))
		if err != nil {
			return fmt.Errorf("failed to decode master key: %w", err)
		}
		if len(key) != 32 {
			return fmt.Errorf("invalid master key length")
		}
		masterKey = key
		return nil
	}

	// Generate new key
	masterKey = make([]byte, 32) // AES-256
	if _, err := io.ReadFull(rand.Reader, masterKey); err != nil {
		return fmt.Errorf("failed to generate master key: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Save key (hex encoded)
	if err := os.WriteFile(keyPath, []byte(hex.EncodeToString(masterKey)), 0600); err != nil {
		return fmt.Errorf("failed to save master key: %w", err)
	}

	return nil
}

// Encrypt encrypts plaintext using AES-GCM
func Encrypt(plaintext string) (string, error) {
	if len(masterKey) == 0 {
		return "", fmt.Errorf("encryption not initialized")
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-GCM
func Decrypt(ciphertext string) (string, error) {
	if len(masterKey) == 0 {
		return "", fmt.Errorf("encryption not initialized")
	}

	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptIfNotEmpty encrypts only non-empty strings
func EncryptIfNotEmpty(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	return Encrypt(plaintext)
}

// DecryptIfNotEmpty decrypts only non-empty strings
func DecryptIfNotEmpty(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	return Decrypt(ciphertext)
}
