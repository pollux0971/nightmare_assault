package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"
)

const (
	// EncryptedPrefix marks an encrypted value
	EncryptedPrefix = "encrypted:AES256:"
)

var (
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	ErrDecryptionFailed  = errors.New("decryption failed")
)

// getEncryptionKey derives an encryption key from the machine ID or a fallback.
// This provides basic protection - the key is tied to the machine.
func getEncryptionKey() ([]byte, error) {
	// Try to read machine-id (Linux/systemd)
	machineID, err := os.ReadFile("/etc/machine-id")
	if err == nil && len(machineID) > 0 {
		hash := sha256.Sum256(append(machineID, []byte("nightmare-assault-v1")...))
		return hash[:], nil
	}

	// Try /var/lib/dbus/machine-id (older Linux)
	machineID, err = os.ReadFile("/var/lib/dbus/machine-id")
	if err == nil && len(machineID) > 0 {
		hash := sha256.Sum256(append(machineID, []byte("nightmare-assault-v1")...))
		return hash[:], nil
	}

	// Fallback: use hostname + user
	hostname, _ := os.Hostname()
	homeDir, _ := os.UserHomeDir()
	fallback := hostname + ":" + homeDir + ":nightmare-assault-v1"
	hash := sha256.Sum256([]byte(fallback))
	return hash[:], nil
}

// Encrypt encrypts plaintext using AES-256-GCM.
func Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	key, err := getEncryptionKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	encoded := base64.StdEncoding.EncodeToString(ciphertext)

	return EncryptedPrefix + encoded, nil
}

// Decrypt decrypts ciphertext that was encrypted with Encrypt.
func Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Check for prefix
	if !strings.HasPrefix(ciphertext, EncryptedPrefix) {
		// Not encrypted, return as-is (for backward compatibility)
		return ciphertext, nil
	}

	encoded := strings.TrimPrefix(ciphertext, EncryptedPrefix)
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	key, err := getEncryptionKey()
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

// IsEncrypted checks if a value is encrypted.
func IsEncrypted(value string) bool {
	return strings.HasPrefix(value, EncryptedPrefix)
}

// EncryptAPIKey encrypts an API key and stores it in the config.
func (c *Config) EncryptAPIKey(providerID, apiKey string) error {
	encrypted, err := Encrypt(apiKey)
	if err != nil {
		return err
	}
	c.API.APIKeys[providerID] = encrypted
	return nil
}

// DecryptAPIKey retrieves and decrypts an API key from the config.
func (c *Config) DecryptAPIKey(providerID string) (string, error) {
	encrypted, ok := c.API.APIKeys[providerID]
	if !ok {
		return "", nil
	}
	return Decrypt(encrypted)
}

// MaskAPIKey returns a masked version of an API key for display.
func MaskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "****"
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
}
