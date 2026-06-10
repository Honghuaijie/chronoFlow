package security

import "testing"

func TestTokenCryptoEncryptDecrypt_RoundTrip(t *testing.T) {
	crypto, err := NewTokenCrypto("12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("NewTokenCrypto() error = %v", err)
	}

	ciphertext, err := crypto.Encrypt("executor-secret")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if ciphertext == "" {
		t.Fatal("Encrypt() returned empty ciphertext")
	}
	if ciphertext == "executor-secret" {
		t.Fatal("Encrypt() returned plaintext")
	}

	plaintext, err := crypto.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if plaintext != "executor-secret" {
		t.Fatalf("Decrypt() = %q, want %q", plaintext, "executor-secret")
	}
}

func TestTokenCryptoRejectsInvalidKeyLength(t *testing.T) {
	if _, err := NewTokenCrypto("short"); err == nil {
		t.Fatal("NewTokenCrypto() error = nil, want invalid key error")
	}
}

func TestTokenCryptoDecryptRejectsInvalidCiphertext(t *testing.T) {
	crypto, err := NewTokenCrypto("12345678901234567890123456789012")
	if err != nil {
		t.Fatalf("NewTokenCrypto() error = %v", err)
	}

	if _, err := crypto.Decrypt("not-base64"); err == nil {
		t.Fatal("Decrypt() error = nil, want invalid ciphertext error")
	}
}
