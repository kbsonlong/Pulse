package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"Pulse/internal/models"
)

// EncryptionService 加密服务接口
type EncryptionService interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
	EncryptDataSourceConfig(config *models.DataSourceConfig) error
	DecryptDataSourceConfig(config *models.DataSourceConfig) error
}

// aesEncryptionService AES加密服务实现
type aesEncryptionService struct {
	key []byte
}

// NewAESEncryptionService 创建AES加密服务
func NewAESEncryptionService(key string) EncryptionService {
	// 确保密钥长度为32字节（AES-256）
	keyBytes := []byte(key)
	if len(keyBytes) < 32 {
		// 如果密钥不足32字节，用0填充
		padded := make([]byte, 32)
		copy(padded, keyBytes)
		keyBytes = padded
	} else if len(keyBytes) > 32 {
		// 如果密钥超过32字节，截取前32字节
		keyBytes = keyBytes[:32]
	}
	
	return &aesEncryptionService{
		key: keyBytes,
	}
}

// Encrypt 加密字符串
func (s *aesEncryptionService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}
	
	plaintextBytes := []byte(plaintext)
	ciphertext := make([]byte, aes.BlockSize+len(plaintextBytes))
	iv := ciphertext[:aes.BlockSize]
	
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintextBytes)
	
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt 解密字符串
func (s *aesEncryptionService) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	
	ciphertextBytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	
	if len(ciphertextBytes) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}
	
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}
	
	iv := ciphertextBytes[:aes.BlockSize]
	ciphertextBytes = ciphertextBytes[aes.BlockSize:]
	
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertextBytes, ciphertextBytes)
	
	return string(ciphertextBytes), nil
}

// EncryptDataSourceConfig 加密数据源配置中的敏感信息
func (s *aesEncryptionService) EncryptDataSourceConfig(config *models.DataSourceConfig) error {
	if config == nil {
		return nil
	}
	
	// 加密密码
	if config.Password != nil && *config.Password != "" {
		encrypted, err := s.Encrypt(*config.Password)
		if err != nil {
			return err
		}
		config.Password = &encrypted
	}
	
	// 加密Token
	if config.Token != nil && *config.Token != "" {
		encrypted, err := s.Encrypt(*config.Token)
		if err != nil {
			return err
		}
		config.Token = &encrypted
	}
	
	return nil
}

// DecryptDataSourceConfig 解密数据源配置中的敏感信息
func (s *aesEncryptionService) DecryptDataSourceConfig(config *models.DataSourceConfig) error {
	if config == nil {
		return nil
	}
	
	// 解密密码
	if config.Password != nil && *config.Password != "" {
		decrypted, err := s.Decrypt(*config.Password)
		if err != nil {
			return err
		}
		config.Password = &decrypted
	}
	
	// 解密Token
	if config.Token != nil && *config.Token != "" {
		decrypted, err := s.Decrypt(*config.Token)
		if err != nil {
			return err
		}
		config.Token = &decrypted
	}
	
	return nil
}