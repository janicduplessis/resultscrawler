package lib

// The Crypto interface exposes various cryto algorithms and utilities.
type Crypto interface {
	AESEncrypt(data []byte) ([]byte, error)
	AESDecrypt(data []byte) ([]byte, error)
	GenerateRandomKey(strength int) []byte
}
