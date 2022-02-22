package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"github.com/teitiago/task-manager-poc/internal/config"
)

type aesEncryption struct {
	block cipher.Block
}

// NewAESEncryption Creates a new instance of the aes encryption.
// It uses a secret to serve as a symmetric key.
// This uses counter mode https://en.wikipedia.org/wiki/Galois/Counter_Mode
func NewAESEncryption() *aesEncryption {
	block, err := aes.NewCipher([]byte(config.GetEnv("AES_SECRET", "")))
	if err != nil {
		panic(err)
	}
	return &aesEncryption{block: block}
}

// Encrypt encrypts a given string. If something goes wrong an error is returned.
func (e *aesEncryption) Encrypt(text string) (string, error) {

	gcm, err := cipher.NewGCM(e.block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	// populates nonce with a cryptographically secure random sequence
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	encryptedByte := gcm.Seal(nonce, nonce, []byte(text), nil)
	return string(encryptedByte), nil

}

// Decrypt decrypts a given encoded message based on the aes secret.
func (e *aesEncryption) Decrypt(cipherText string) (string, error) {
	gcm, err := cipher.NewGCM(e.block)
	if err != nil {
		return "", err
	}

	cipherBytes := []byte(cipherText)
	nonceSize := gcm.NonceSize()
	if len(cipherBytes) < nonceSize {
		return "", err
	}

	nonce, cipherBytes := cipherBytes[:nonceSize], cipherBytes[nonceSize:]
	text, err := gcm.Open(nil, nonce, cipherBytes, nil)
	if err != nil {
		return "", err
	}
	return string(text), nil
}
