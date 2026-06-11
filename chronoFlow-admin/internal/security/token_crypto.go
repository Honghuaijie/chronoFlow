package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"

	"chronoFlow-admin/internal/biz"
	"chronoFlow-admin/internal/conf"

	"github.com/google/wire"
)

const tokenEncryptKeySize = 32

var ProviderSet = wire.NewSet(
	NewTokenCryptoFromConf,
	wire.Bind(new(biz.TokenCipher), new(*TokenCrypto)),
)

type TokenCrypto struct {
	aead cipher.AEAD
}

func NewTokenCryptoFromConf(c *conf.Security) (*TokenCrypto, error) {
	if c == nil {
		return nil, fmt.Errorf("security config is required")
	}
	return NewTokenCrypto(c.TokenEncryptKey)
}

func NewTokenCrypto(key string) (*TokenCrypto, error) {
	if len(key) != tokenEncryptKeySize {
		return nil, fmt.Errorf("token encrypt key must be %d bytes", tokenEncryptKeySize)
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &TokenCrypto{aead: aead}, nil
}

func (c *TokenCrypto) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := c.aead.Seal(nil, nonce, []byte(plaintext), nil)
	payload := append(nonce, sealed...)
	return base64.StdEncoding.EncodeToString(payload), nil
}

func (c *TokenCrypto) Decrypt(ciphertext string) (string, error) {
	payload, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	nonceSize := c.aead.NonceSize()
	if len(payload) <= nonceSize {
		return "", fmt.Errorf("invalid token ciphertext")
	}
	nonce := payload[:nonceSize]
	sealed := payload[nonceSize:]
	plaintext, err := c.aead.Open(nil, nonce, sealed, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
