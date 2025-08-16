package crypto

import (
	"testing"
	"strings"
)

func TestNewAESEncryptor(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "valid 32-byte key",
			key:     "12345678901234567890123456789012",
			wantErr: false,
		},
		{
			name:    "invalid short key",
			key:     "short",
			wantErr: true,
		},
		{
			name:    "invalid long key",
			key:     "123456789012345678901234567890123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAESEncryptor(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAESEncryptor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAESEncryptor_EncryptDecrypt(t *testing.T) {
	key := "12345678901234567890123456789012"
	encryptor, err := NewAESEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "simple text",
			plaintext: "hello world",
		},
		{
			name:      "empty string",
			plaintext: "",
		},
		{
			name:      "password with special chars",
			plaintext: "P@ssw0rd!@#$%^&*()",
		},
		{
			name:      "long text",
			plaintext: strings.Repeat("abcdefghijklmnopqrstuvwxyz", 10),
		},
		{
			name:      "unicode text",
			plaintext: "你好世界🌍",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 加密
			ciphertext, err := encryptor.Encrypt(tt.plaintext)
			if err != nil {
				t.Errorf("Encrypt() error = %v", err)
				return
			}

			// 空字符串应该返回空字符串
			if tt.plaintext == "" {
				if ciphertext != "" {
					t.Errorf("Expected empty ciphertext for empty plaintext, got %s", ciphertext)
				}
				return
			}

			// 密文不应该等于明文
			if ciphertext == tt.plaintext {
				t.Errorf("Ciphertext should not equal plaintext")
			}

			// 解密
			decrypted, err := encryptor.Decrypt(ciphertext)
			if err != nil {
				t.Errorf("Decrypt() error = %v", err)
				return
			}

			// 解密后应该等于原文
			if decrypted != tt.plaintext {
				t.Errorf("Decrypted text = %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestAESEncryptor_DecryptInvalidData(t *testing.T) {
	key := "12345678901234567890123456789012"
	encryptor, err := NewAESEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	tests := []struct {
		name       string
		ciphertext string
		wantErr    bool
	}{
		{
			name:       "invalid base64",
			ciphertext: "invalid-base64!",
			wantErr:    true,
		},
		{
			name:       "too short data",
			ciphertext: "YWJj", // "abc" in base64, too short
			wantErr:    true,
		},
		{
			name:       "empty string",
			ciphertext: "",
			wantErr:    false, // should return empty string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := encryptor.Decrypt(tt.ciphertext)
			if (err != nil) != tt.wantErr {
				t.Errorf("Decrypt() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsEncrypted(t *testing.T) {
	key := "12345678901234567890123456789012"
	encryptor, err := NewAESEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	// 加密一些数据
	encryptedData, err := encryptor.Encrypt("test data")
	if err != nil {
		t.Fatalf("Failed to encrypt test data: %v", err)
	}

	tests := []struct {
		name string
		data string
		want bool
	}{
		{
			name: "encrypted data",
			data: encryptedData,
			want: true,
		},
		{
			name: "plain text",
			data: "plain text",
			want: false,
		},
		{
			name: "empty string",
			data: "",
			want: false,
		},
		{
			name: "invalid base64",
			data: "invalid-base64!",
			want: false,
		},
		{
			name: "short base64",
			data: "YWJj", // "abc" in base64
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsEncrypted(tt.data); got != tt.want {
				t.Errorf("IsEncrypted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAESEncryptor_ConsistentEncryption(t *testing.T) {
	key := "12345678901234567890123456789012"
	encryptor, err := NewAESEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	plaintext := "test data for consistency"

	// 多次加密同一数据，结果应该不同（因为使用随机nonce）
	ciphertext1, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("First encryption failed: %v", err)
	}

	ciphertext2, err := encryptor.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Second encryption failed: %v", err)
	}

	// 密文应该不同
	if ciphertext1 == ciphertext2 {
		t.Error("Multiple encryptions of same data should produce different ciphertexts")
	}

	// 但解密结果应该相同
	decrypted1, err := encryptor.Decrypt(ciphertext1)
	if err != nil {
		t.Fatalf("First decryption failed: %v", err)
	}

	decrypted2, err := encryptor.Decrypt(ciphertext2)
	if err != nil {
		t.Fatalf("Second decryption failed: %v", err)
	}

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Errorf("Decryption failed: got %s and %s, want %s", decrypted1, decrypted2, plaintext)
	}
}