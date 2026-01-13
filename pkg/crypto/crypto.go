package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"sync"
)

var (
	ErrKeyNotSet      = errors.New("ENCRYPTION_KEY environment variable not set")
	ErrKeyTooShort    = errors.New("ENCRYPTION_KEY must be at least 32 bytes")
	ErrInvalidData    = errors.New("invalid encrypted data")
	ErrDecryptionFail = errors.New("decryption failed")
)

var (
	gcm     cipher.AEAD
	gcmOnce sync.Once
	gcmErr  error
)

// initGCM initializes the AES-GCM cipher once
func initGCM() {
	gcmOnce.Do(func() {
		key := os.Getenv("ENCRYPTION_KEY")
		if key == "" {
			// Fallback to JWT_SECRET if ENCRYPTION_KEY not set
			key = os.Getenv("JWT_SECRET")
		}
		if key == "" {
			gcmErr = ErrKeyNotSet
			return
		}

		// Ensure key is exactly 32 bytes for AES-256
		keyBytes := []byte(key)
		if len(keyBytes) < 32 {
			gcmErr = ErrKeyTooShort
			return
		}
		keyBytes = keyBytes[:32] // Use first 32 bytes

		block, err := aes.NewCipher(keyBytes)
		if err != nil {
			gcmErr = err
			return
		}

		gcm, gcmErr = cipher.NewGCM(block)
	})
}

// Encrypt encrypts plaintext using AES-256-GCM and returns base64-encoded ciphertext
// The nonce is prepended to the ciphertext before encoding
func Encrypt(plaintext string) (string, error) {
	initGCM()
	if gcmErr != nil {
		return "", gcmErr
	}

	// Create a unique nonce for this encryption
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt and append nonce + ciphertext
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode as base64 for safe database storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded ciphertext that was encrypted with Encrypt()
func Decrypt(ciphertext string) (string, error) {
	initGCM()
	if gcmErr != nil {
		return "", gcmErr
	}

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", ErrInvalidData
	}

	// Extract nonce from the beginning
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidData
	}

	nonce, encryptedData := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", ErrDecryptionFail
	}

	return string(plaintext), nil
}
