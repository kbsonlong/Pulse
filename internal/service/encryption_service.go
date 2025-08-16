package service

import (
	"fmt"
	"os"
	"strings"

	"pulse/internal/models"
	"pulse/internal/pkg/crypto"
)

// EncryptionService 加密服务接口
type EncryptionService interface {
	// EncryptDataSourceConfig 加密数据源配置
	EncryptDataSourceConfig(config *models.DataSourceConfig) error
	// DecryptDataSourceConfig 解密数据源配置
	DecryptDataSourceConfig(config *models.DataSourceConfig) error
	// IsConfigEncrypted 检查配置是否已加密
	IsConfigEncrypted(config *models.DataSourceConfig) bool
}

// encryptionService 加密服务实现
type encryptionService struct {
	encryptor crypto.Encryptor
}

// NewEncryptionService 创建新的加密服务实例
func NewEncryptionService() (EncryptionService, error) {
	// 从环境变量获取加密密钥
	key := os.Getenv("DATASOURCE_ENCRYPTION_KEY")
	if key == "" {
		// 如果没有设置环境变量，使用默认密钥（生产环境中应该设置环境变量）
		key = "12345678901234567890123456789012" // 32字节密钥
	}

	encryptor, err := crypto.NewAESEncryptor(key)
	if err != nil {
		return nil, fmt.Errorf("创建加密器失败: %w", err)
	}

	return &encryptionService{
		encryptor: encryptor,
	}, nil
}

// EncryptDataSourceConfig 加密数据源配置中的敏感信息
func (s *encryptionService) EncryptDataSourceConfig(config *models.DataSourceConfig) error {
	if config == nil {
		return nil
	}

	// 加密密码
	if config.Password != nil && *config.Password != "" && !crypto.IsEncrypted(*config.Password) {
		encryptedPassword, err := s.encryptor.Encrypt(*config.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %w", err)
		}
		config.Password = &encryptedPassword
	}

	// 加密Token
	if config.Token != nil && *config.Token != "" && !crypto.IsEncrypted(*config.Token) {
		encryptedToken, err := s.encryptor.Encrypt(*config.Token)
		if err != nil {
			return fmt.Errorf("failed to encrypt token: %w", err)
		}
		config.Token = &encryptedToken
	}

	// 加密Headers中的敏感信息
	for key, value := range config.Headers {
		if s.isSensitiveKey(key) && !crypto.IsEncrypted(value) {
			encryptedValue, err := s.encryptor.Encrypt(value)
			if err != nil {
				return fmt.Errorf("failed to encrypt header %s: %w", key, err)
			}
			config.Headers[key] = encryptedValue
		}
	}

	// 加密Parameters中的敏感信息
	for key, value := range config.Parameters {
		if s.isSensitiveKey(key) {
			if strValue, ok := value.(string); ok && !crypto.IsEncrypted(strValue) {
				encryptedValue, err := s.encryptor.Encrypt(strValue)
				if err != nil {
					return fmt.Errorf("failed to encrypt parameter %s: %w", key, err)
				}
				config.Parameters[key] = encryptedValue
			}
		}
	}

	return nil
}

// DecryptDataSourceConfig 解密数据源配置中的敏感信息
func (s *encryptionService) DecryptDataSourceConfig(config *models.DataSourceConfig) error {
	if config == nil {
		return nil
	}

	// 解密密码
	if config.Password != nil && *config.Password != "" && crypto.IsEncrypted(*config.Password) {
		decryptedPassword, err := s.encryptor.Decrypt(*config.Password)
		if err != nil {
			return fmt.Errorf("failed to decrypt password: %w", err)
		}
		config.Password = &decryptedPassword
	}

	// 解密Token
	if config.Token != nil && *config.Token != "" && crypto.IsEncrypted(*config.Token) {
		decryptedToken, err := s.encryptor.Decrypt(*config.Token)
		if err != nil {
			return fmt.Errorf("failed to decrypt token: %w", err)
		}
		config.Token = &decryptedToken
	}

	// 解密Headers中的敏感信息
	for key, value := range config.Headers {
		if s.isSensitiveKey(key) && crypto.IsEncrypted(value) {
			decryptedValue, err := s.encryptor.Decrypt(value)
			if err != nil {
				return fmt.Errorf("failed to decrypt header %s: %w", key, err)
			}
			config.Headers[key] = decryptedValue
		}
	}

	// 解密Parameters中的敏感信息
	for key, value := range config.Parameters {
		if s.isSensitiveKey(key) {
			if strValue, ok := value.(string); ok && crypto.IsEncrypted(strValue) {
				decryptedValue, err := s.encryptor.Decrypt(strValue)
				if err != nil {
					return fmt.Errorf("failed to decrypt parameter %s: %w", key, err)
				}
				config.Parameters[key] = decryptedValue
			}
		}
	}

	return nil
}

// IsConfigEncrypted 检查配置是否已加密
func (s *encryptionService) IsConfigEncrypted(config *models.DataSourceConfig) bool {
	if config == nil {
		return false
	}

	// 检查密码是否加密
	if config.Password != nil && *config.Password != "" && crypto.IsEncrypted(*config.Password) {
		return true
	}

	// 检查Token是否加密
	if config.Token != nil && *config.Token != "" && crypto.IsEncrypted(*config.Token) {
		return true
	}

	// 检查Headers中的敏感信息是否加密
	for key, value := range config.Headers {
		if s.isSensitiveKey(key) && crypto.IsEncrypted(value) {
			return true
		}
	}

	// 检查Parameters中的敏感信息是否加密
	for key, value := range config.Parameters {
		if s.isSensitiveKey(key) {
			if strValue, ok := value.(string); ok && crypto.IsEncrypted(strValue) {
				return true
			}
		}
	}

	return false
}

// isSensitiveKey 检查字段名是否包含敏感关键词
func (s *encryptionService) isSensitiveKey(fieldName string) bool {
	fieldLower := strings.ToLower(fieldName)
	sensitiveKeywords := []string{"password", "token", "secret", "key", "auth", "credential", "pass", "api_key", "apikey", "access_token", "refresh_token"}
	
	for _, keyword := range sensitiveKeywords {
		if strings.Contains(fieldLower, keyword) {
			return true
		}
	}
	return false
}