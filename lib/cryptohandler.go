package lib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"

	"code.google.com/p/go.crypto/bcrypt"
)

// CryptoHandler is an utility for various crypto algorithms.
type CryptoHandler struct {
	block cipher.Block
}

// NewCryptoHandler creates a new CryptoHandler object.
// Key must be 16 bytes long.
func NewCryptoHandler(key string) *CryptoHandler {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		panic(err)
	}

	return &CryptoHandler{
		block: block,
	}
}

// AESEncrypt encrypts the data using AES-128.
func (hndl *CryptoHandler) AESEncrypt(data []byte) ([]byte, error) {
	iv := hndl.GenerateRandomKey(aes.BlockSize)
	if iv == nil {
		return nil, errors.New("Failed to generate random iv")
	}
	encrypter := cipher.NewCFBEncrypter(hndl.block, iv)
	encrypted := make([]byte, len(data))
	encrypter.XORKeyStream(encrypted, data)
	return append(iv, encrypted...), nil
}

// AESDecrypt decrypts the data using AES-128.
func (hndl *CryptoHandler) AESDecrypt(data []byte) ([]byte, error) {
	size := aes.BlockSize
	if len(data) <= size {
		return nil, errors.New("Decryption failed, data does not contain iv")
	}
	iv := data[:size]
	data = data[size:]
	decrypter := cipher.NewCFBDecrypter(hndl.block, iv)
	decrypted := make([]byte, len(data))
	decrypter.XORKeyStream(decrypted, data)
	return decrypted, nil
}

// GenerateRandomKey creates a random key with the given strength.
func (hndl *CryptoHandler) GenerateRandomKey(strength int) []byte {
	k := make([]byte, strength)
	if _, err := rand.Read(k); err != nil {
		return nil
	}
	return k
}

func (hndl *CryptoHandler) CompareHashAndPassword(hash string, password string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (hndl *CryptoHandler) GenerateFromPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
