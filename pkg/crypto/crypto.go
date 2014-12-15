package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"log"

	"code.google.com/p/go.crypto/bcrypt"
)

var block cipher.Block

// Init initializes the crypto package.
// Key must be 16 bytes long.
func Init(key string) {
	block = mustCreateBlock([]byte(key))
}

// AESEncrypt encrypts the data using AES-128.
func AESEncrypt(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}
	iv := GenerateRandomKey(aes.BlockSize)
	if iv == nil {
		return nil, errors.New("Failed to generate random iv")
	}
	encrypter := cipher.NewCFBEncrypter(block, iv)
	encrypted := make([]byte, len(data))
	encrypter.XORKeyStream(encrypted, data)
	return append(iv, encrypted...), nil
}

// AESDecrypt decrypts the data using AES-128.
func AESDecrypt(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}
	size := aes.BlockSize
	if len(data) <= size {
		return nil, errors.New("Decryption failed, data does not contain iv")
	}
	iv := data[:size]
	data = data[size:]
	decrypter := cipher.NewCFBDecrypter(block, iv)
	decrypted := make([]byte, len(data))
	decrypter.XORKeyStream(decrypted, data)
	return decrypted, nil
}

// GenerateRandomKey creates a random key with the given strength.
func GenerateRandomKey(strength int) []byte {
	k := make([]byte, strength)
	if _, err := rand.Read(k); err != nil {
		return nil
	}
	return k
}

// CompareHashAndPassword checks if the hashed password matches the plain text one.
func CompareHashAndPassword(hash string, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GenerateFromPassword creates a hash from the plain text password.
func GenerateFromPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func mustCreateBlock(key []byte) cipher.Block {
	b, err := aes.NewCipher([]byte(key))
	if err != nil {
		log.Fatal(err)
	}
	return b
}
