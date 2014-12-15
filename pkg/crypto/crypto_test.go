package crypto

import "testing"

const (
	testAESKey = "1234abcd1234abcd"
)

func TestCryptoEncryptDecrypt(t *testing.T) {
	Init(testAESKey)

	data := []string{
		"H",
		"wdaawdfsefseawdf",
		"H23dawd2@dawd~!",
		"awdoj23389duawdlkjfoeij398uawdouh23e98hawdujho289ahwdoawjdh289hawoudih28d",
	}

	// Test if the encrypting and decrypting gives the same string
	for _, curStr := range data {
		crypted, err := AESEncrypt([]byte(curStr))
		if err != nil {
			t.Errorf("Error encrypting %s. Err: %s", curStr, err.Error())
			continue
		}
		cryptedStr := string(crypted)
		decrypted, err := AESDecrypt([]byte(cryptedStr))
		if err != nil {
			t.Errorf("Error decrypting %s. Err: %s", curStr, err.Error())
		}
		decryptedStr := string(decrypted)
		if decryptedStr != curStr {
			t.Errorf("Decrypted data doesn't match initial. Init: %s, Decrypted: %s", curStr, decryptedStr)
		}
	}
}

func TestCryptoEncrypt(t *testing.T) {
	Init(testAESKey)
	// Test if encrypting the same data gives different encrypted data
	data := []byte("TEST")
	crypted1, err := AESEncrypt(data)
	if err != nil {
		t.Errorf("Error decrypting %s. Err: %s", string(data), err.Error())
	}
	crypted2, err := AESEncrypt(data)
	if string(crypted1) == string(crypted2) {
		t.Errorf("Encrypted data is the same for the same data. Crypted1: %s, Crypted2: %s", string(crypted1), string(crypted2))
	}
}
