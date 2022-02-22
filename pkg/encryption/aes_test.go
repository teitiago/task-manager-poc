//go:build unit

package encryption

import (
	"crypto/aes"
	"math/rand"
	"os"
	"testing"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// TestValidAes Validates that the encryption and decryption process works as expected.
func TestValidAes(t *testing.T) {

	shortLen := 10
	longLen := 3000

	testingMap := []struct {
		name      string
		key       string
		inputText string
	}{

		{
			name:      "valid 128-bit key",
			key:       "XnZr4u7x!A%D*G-K",
			inputText: randStringRunes(shortLen),
		},
		{
			name:      "valid 128-bit key long",
			key:       "XnZr4u7x!A%D*G-K",
			inputText: randStringRunes(longLen),
		},
		{
			name:      "valid 256-bit key",
			key:       "D*G-KaNdRgUkXp2s5v8y/B?E(H+MbQeS",
			inputText: randStringRunes(shortLen),
		},
		{
			name:      "valid 256-bit key long",
			key:       "D*G-KaNdRgUkXp2s5v8y/B?E(H+MbQeS",
			inputText: randStringRunes(longLen),
		},
	}

	for _, test := range testingMap {
		_ = os.Setenv("AES_SECRET", test.key)
		aes := NewAESEncryption()
		encrypted, err := aes.Encrypt(test.inputText)
		if err != nil {
			t.Fatalf("unexpected test error, %v", err.Error())
		}
		if encrypted == "" {
			t.Fatal("unexpected empty output")
		}
		decrypted, err := aes.Decrypt(encrypted)
		if err != nil {
			t.Fatalf("unexpected test error, %v", err.Error())
		}
		if decrypted != test.inputText {
			t.Fatalf("expected %v got %v", test.inputText, decrypted)
		}
	}

}

// TestInvalidAes Validates that decrypted and encrypted don't match
func TestInvalidAes(t *testing.T) {
	t.Run("change block and get different decryption", func(t *testing.T) {
		_ = os.Setenv("AES_SECRET", "(G-KaPdSgVkYp3s6v9y$B&E)H@MbQeTh")
		aesM := NewAESEncryption()
		ecnrypted, err := aesM.Encrypt("test")
		if err != nil {
			t.Fatalf("unexpected error, %v", err.Error())
		}
		aesM.block, _ = aes.NewCipher([]byte("bPeShVmYq3t6v9y$"))
		decrypted, _ := aesM.Decrypt(ecnrypted)
		if decrypted == "test" {
			t.Errorf("expecting different values")
		}
	})
}
