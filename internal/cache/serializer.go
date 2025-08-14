package cache

import (
	"encoding/json"
	"fmt"
)

// JSONSerializer JSON序列化器
type JSONSerializer struct{}

// Serialize 序列化
func (s *JSONSerializer) Serialize(v interface{}) ([]byte, error) {
	switch val := v.(type) {
	case string:
		return []byte(val), nil
	case []byte:
		return val, nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return []byte(fmt.Sprintf("%d", val)), nil
	case float32, float64:
		return []byte(fmt.Sprintf("%f", val)), nil
	case bool:
		if val {
			return []byte("true"), nil
		}
		return []byte("false"), nil
	default:
		return json.Marshal(v)
	}
}

// Deserialize 反序列化
func (s *JSONSerializer) Deserialize(data []byte, v interface{}) error {
	switch val := v.(type) {
	case *string:
		*val = string(data)
		return nil
	case *[]byte:
		*val = data
		return nil
	default:
		return json.Unmarshal(data, v)
	}
}

// NoOpSerializer 无操作序列化器（直接存储字符串）
type NoOpSerializer struct{}

// Serialize 序列化
func (s *NoOpSerializer) Serialize(v interface{}) ([]byte, error) {
	switch val := v.(type) {
	case string:
		return []byte(val), nil
	case []byte:
		return val, nil
	default:
		return nil, fmt.Errorf("NoOpSerializer only supports string and []byte types")
	}
}

// Deserialize 反序列化
func (s *NoOpSerializer) Deserialize(data []byte, v interface{}) error {
	switch val := v.(type) {
	case *string:
		*val = string(data)
		return nil
	case *[]byte:
		*val = data
		return nil
	default:
		return fmt.Errorf("NoOpSerializer only supports *string and *[]byte types")
	}
}