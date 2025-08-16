package service

import (
	"os"
	"testing"

	"pulse/internal/models"
	"pulse/internal/pkg/crypto"
)

// stringPtr 返回字符串的指针
func stringPtr(s string) *string {
	return &s
}

func TestNewEncryptionService(t *testing.T) {
	// 测试使用环境变量
	os.Setenv("DATASOURCE_ENCRYPTION_KEY", "12345678901234567890123456789012")
	defer os.Unsetenv("DATASOURCE_ENCRYPTION_KEY")

	service, err := NewEncryptionService()
	if err != nil {
		t.Fatalf("NewEncryptionService() error = %v", err)
	}
	if service == nil {
		t.Fatal("NewEncryptionService() returned nil service")
	}
}

func TestNewEncryptionService_DefaultKey(t *testing.T) {
	// 测试使用默认密钥
	os.Unsetenv("DATASOURCE_ENCRYPTION_KEY")

	service, err := NewEncryptionService()
	if err != nil {
		t.Fatalf("NewEncryptionService() error = %v", err)
	}
	if service == nil {
		t.Fatal("NewEncryptionService() returned nil service")
	}
}

func TestEncryptionService_EncryptDecryptDataSourceConfig(t *testing.T) {
	service, err := NewEncryptionService()
	if err != nil {
		t.Fatalf("Failed to create encryption service: %v", err)
	}

	tests := []struct {
		name   string
		config *models.DataSourceConfig
	}{
		{
			name: "basic config with password and token",
			config: &models.DataSourceConfig{
				URL:      "http://localhost:8080",
				Username: stringPtr("admin"),
				Password: stringPtr("secret123"),
				Token:    stringPtr("token123"),
				Database: stringPtr("testdb"),
			},
		},
		{
			name: "config with sensitive headers and parameters",
			config: &models.DataSourceConfig{
				URL:      "http://localhost:8080",
				Username: stringPtr("admin"),
				Password: stringPtr("secret123"),
				Headers: map[string]string{
					"Authorization": "Bearer token123",
					"X-API-Key":     "key123",
					"Content-Type":  "application/json",
				},
				Parameters: map[string]interface{}{
					"client_secret": "secret456",
					"timeout":       30,
					"host":          "localhost",
				},
			},
		},
		{
			name: "empty config",
			config: &models.DataSourceConfig{
				URL:      "http://localhost:8080",
				Username: stringPtr("admin"),
				Database: stringPtr("testdb"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 保存原始值
			var originalPassword, originalToken string
			if tt.config.Password != nil {
				originalPassword = *tt.config.Password
			}
			if tt.config.Token != nil {
				originalToken = *tt.config.Token
			}
			originalHeaders := make(map[string]string)
			for k, v := range tt.config.Headers {
				originalHeaders[k] = v
			}
			originalParameters := make(map[string]interface{})
			for k, v := range tt.config.Parameters {
				originalParameters[k] = v
			}

			// 加密
			err := service.EncryptDataSourceConfig(tt.config)
			if err != nil {
				t.Errorf("EncryptDataSourceConfig() error = %v", err)
				return
			}

			// 验证敏感字段已加密
			if originalPassword != "" {
				if tt.config.Password != nil && *tt.config.Password == originalPassword {
					t.Error("Password should be encrypted")
				}
				if tt.config.Password != nil && !crypto.IsEncrypted(*tt.config.Password) {
					t.Error("Password should be marked as encrypted")
				}
			}

			if originalToken != "" {
				if tt.config.Token != nil && *tt.config.Token == originalToken {
					t.Error("Token should be encrypted")
				}
				if tt.config.Token != nil && !crypto.IsEncrypted(*tt.config.Token) {
					t.Error("Token should be marked as encrypted")
				}
			}

			// 验证非敏感字段未加密
			if tt.config.URL != "http://localhost:8080" {
				t.Error("URL should not be encrypted")
			}
			if tt.config.Username == nil || *tt.config.Username != "admin" {
				var actual string
				if tt.config.Username != nil {
					actual = *tt.config.Username
				}
				t.Errorf("Username should be admin, got %s", actual)
			}

			// 解密
			err = service.DecryptDataSourceConfig(tt.config)
			if err != nil {
				t.Errorf("DecryptDataSourceConfig() error = %v", err)
				return
			}

			// 验证解密后的值
			if originalPassword != "" {
				if tt.config.Password == nil || *tt.config.Password != originalPassword {
					var actual string
					if tt.config.Password != nil {
						actual = *tt.config.Password
					}
					t.Errorf("Decrypted password = %s, want %s", actual, originalPassword)
				}
			}

			if originalToken != "" {
				if tt.config.Token == nil || *tt.config.Token != originalToken {
					var actual string
					if tt.config.Token != nil {
						actual = *tt.config.Token
					}
					t.Errorf("Decrypted token = %s, want %s", actual, originalToken)
				}
			}

			// 验证Headers字段解密结果
			for key, originalValue := range originalHeaders {
				if currentValue, exists := tt.config.Headers[key]; exists {
					if currentValue != originalValue {
						t.Errorf("Decrypted headers[%s] = %v, want %v", key, currentValue, originalValue)
					}
				}
			}

			// 验证Parameters字段解密结果
			for key, originalValue := range originalParameters {
				if currentValue, exists := tt.config.Parameters[key]; exists {
					if currentValue != originalValue {
						t.Errorf("Decrypted parameters[%s] = %v, want %v", key, currentValue, originalValue)
					}
				}
			}
		})
	}
}

func TestEncryptionService_IsConfigEncrypted(t *testing.T) {
	service, err := NewEncryptionService()
	if err != nil {
		t.Fatalf("Failed to create encryption service: %v", err)
	}

	tests := []struct {
		name   string
		config *models.DataSourceConfig
		want   bool
	}{
		{
			name: "nil config",
			config: nil,
			want: false,
		},
		{
			name: "unencrypted config",
			config: &models.DataSourceConfig{
				Password: stringPtr("plaintext"),
				Token:    stringPtr("plaintoken"),
			},
			want: false,
		},
		{
			name: "已加密的配置",
			config: &models.DataSourceConfig{
				Username: stringPtr("user"),
				Password: stringPtr("dGVzdGVuY3J5cHRlZGRhdGF0aGF0aXNsb25nZW5vdWdodG9wYXNzdGhlbGVuZ3RoY2hlY2s="), // 足够长的base64字符串
				Headers: map[string]string{
					"Authorization": "dGVzdGVuY3J5cHRlZGRhdGF0aGF0aXNsb25nZW5vdWdodG9wYXNzdGhlbGVuZ3RoY2hlY2s=",
				},
				Parameters: map[string]interface{}{
					"api_key": "dGVzdGVuY3J5cHRlZGRhdGF0aGF0aXNsb25nZW5vdWdodG9wYXNzdGhlbGVuZ3RoY2hlY2s=",
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.IsConfigEncrypted(tt.config)
			if got != tt.want {
				t.Errorf("IsConfigEncrypted() = %v, want %v", got, tt.want)
			}
		})
	}

	// 测试真正的加密数据
	t.Run("real encrypted config", func(t *testing.T) {
		config := &models.DataSourceConfig{
			Password: stringPtr("secret123"),
			Token:    stringPtr("token123"),
		}

		// 加密配置
		err := service.EncryptDataSourceConfig(config)
		if err != nil {
			t.Fatalf("EncryptDataSourceConfig() error = %v", err)
		}

		// 检查是否被识别为已加密
		if !service.IsConfigEncrypted(config) {
			t.Error("Encrypted config should be identified as encrypted")
		}
	})
}

func TestEncryptionService_IsSensitiveField(t *testing.T) {
	service := &encryptionService{}

	tests := []struct {
		name      string
		fieldName string
		want      bool
	}{
		{"password", "password", true},
		{"Password", "Password", true},
		{"user_password", "user_password", true},
		{"token", "token", true},
		{"access_token", "access_token", true},
		{"api_key", "api_key", true},
		{"secret", "secret", true},
		{"client_secret", "client_secret", true},
		{"username", "username", false},
		{"url", "url", false},
		{"database", "database", false},
		{"timeout", "timeout", false},
		{"host", "host", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.isSensitiveKey(tt.fieldName)
			if got != tt.want {
				t.Errorf("isSensitiveKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncryptionService_EncryptDecryptNilConfig(t *testing.T) {
	service, err := NewEncryptionService()
	if err != nil {
		t.Fatalf("Failed to create encryption service: %v", err)
	}

	// 测试nil配置
	err = service.EncryptDataSourceConfig(nil)
	if err != nil {
		t.Errorf("EncryptDataSourceConfig(nil) should not return error, got %v", err)
	}

	err = service.DecryptDataSourceConfig(nil)
	if err != nil {
		t.Errorf("DecryptDataSourceConfig(nil) should not return error, got %v", err)
	}
}

func TestEncryptionService_DoubleEncryption(t *testing.T) {
	service, err := NewEncryptionService()
	if err != nil {
		t.Fatalf("Failed to create encryption service: %v", err)
	}

	config := &models.DataSourceConfig{
		Password: stringPtr("secret123"),
	}

	// 第一次加密
	err = service.EncryptDataSourceConfig(config)
	if err != nil {
		t.Fatalf("First encryption failed: %v", err)
	}

	firstEncryption := config.Password

	// 第二次加密（应该不会重复加密）
	err = service.EncryptDataSourceConfig(config)
	if err != nil {
		t.Fatalf("Second encryption failed: %v", err)
	}

	// 密码应该保持不变（不会重复加密）
	if config.Password != firstEncryption {
		t.Error("Password should not be encrypted twice")
	}

	// 解密应该正常工作
	err = service.DecryptDataSourceConfig(config)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if config.Password == nil || *config.Password != "secret123" {
		var actual string
		if config.Password != nil {
			actual = *config.Password
		}
		t.Errorf("Decrypted password = %s, want %s", actual, "secret123")
	}
}