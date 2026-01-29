package config

import (
	"fmt"
	"os"
	"trae-proxy-go/pkg/models"

	"gopkg.in/yaml.v3"
)

// LoadConfig 从文件加载配置
func LoadConfig(configPath string) (*models.Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config models.Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// SaveConfig 保存配置到文件
func SaveConfig(config *models.Config, configPath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// validateConfig 验证配置的有效性
func validateConfig(config *models.Config) error {
	if config.Domain == "" {
		return fmt.Errorf("域名不能为空")
	}

	if len(config.APIs) == 0 {
		return fmt.Errorf("至少需要配置一个API")
	}

	for i, api := range config.APIs {
		if api.Name == "" {
			return fmt.Errorf("API配置[%d]的名称不能为空", i)
		}
		if api.Endpoint == "" {
			return fmt.Errorf("API配置[%d]的endpoint不能为空", i)
		}
		if api.CustomModelID == "" {
			return fmt.Errorf("API配置[%d]的custom_model_id不能为空", i)
		}
		if api.TargetModelID == "" {
			return fmt.Errorf("API配置[%d]的target_model_id不能为空", i)
		}
	}

	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("服务器端口必须在1-65535之间")
	}

	return nil
}

// GetDefaultConfig 获取默认配置
func GetDefaultConfig() *models.Config {
	return &models.Config{
		Domain: "api.openai.com",
		APIs: []models.API{
			{
				Name:          "默认OpenAI API",
				Endpoint:      "https://api.openai.com",
				CustomModelID: "gpt-4",
				TargetModelID: "gpt-4",
				StreamMode:    "",
				Active:        true,
			},
		},
		Server: models.Server{
			Port:  443,
			Debug: true,
		},
	}
}
