package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var (
	ErrInvalidKey       = errors.New("invalid encryption key")
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	ErrDecryptionFailed = errors.New("decryption failed")
)

// Encryptor 加密器接口
type Encryptor interface {
	// Encrypt 加密数据
	Encrypt(plaintext string) (string, error)
	// Decrypt 解密数据
	Decrypt(ciphertext string) (string, error)
}

// AESEncryptor AES加密器
type AESEncryptor struct {
	key []byte
}

// NewAESEncryptor 创建AES加密器
func NewAESEncryptor(key string) (*AESEncryptor, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("%w: key must be 32 bytes", ErrInvalidKey)
	}
	
	return &AESEncryptor{
		key: []byte(key),
	}, nil
}

// Encrypt 加密数据
func (e *AESEncryptor) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	
	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}
	
	// 生成随机nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	
	// 加密数据
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	
	// 返回base64编码的结果
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密数据
func (e *AESEncryptor) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	
	// 解码base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("%w: invalid base64 encoding", ErrInvalidCiphertext)
	}
	
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}
	
	// 创建GCM模式
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}
	
	// 检查数据长度
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("%w: ciphertext too short", ErrInvalidCiphertext)
	}
	
	// 提取nonce和密文
	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	
	// 解密数据
	plaintext, err := gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}
	
	return string(plaintext), nil
}

// IsEncrypted 检查字符串是否已加密（简单检查base64格式）
func IsEncrypted(data string) bool {
	if data == "" {
		return false
	}
	
	// 尝试解码base64，如果成功且长度合理，可能是加密数据
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return false
	}
	
	// 加密数据至少包含nonce（12字节）+ 密文（至少1字节）+ tag（16字节）
	return len(decoded) >= 29
}