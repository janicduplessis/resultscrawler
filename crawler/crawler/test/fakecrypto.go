package test

type FakeCrypto struct {
}

func (c *FakeCrypto) AESEncrypt(data []byte) ([]byte, error) {
	return data, nil
}

func (c *FakeCrypto) AESDecrypt(data []byte) ([]byte, error) {
	return data, nil
}

func (c *FakeCrypto) GenerateRandomKey(strength int) []byte {
	return []byte("1234")
}
