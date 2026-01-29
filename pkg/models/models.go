package models

// API 配置结构
type API struct {
	Name          string `yaml:"name" json:"name"`
	Endpoint      string `yaml:"endpoint" json:"endpoint"`
	CustomModelID string `yaml:"custom_model_id" json:"custom_model_id"`
	TargetModelID string `yaml:"target_model_id" json:"target_model_id"`
	StreamMode    string `yaml:"stream_mode" json:"stream_mode"` // "true", "false", or null
	Active        bool   `yaml:"active" json:"active"`
}

// Server 配置结构
type Server struct {
	Port  int  `yaml:"port" json:"port"`
	Debug bool `yaml:"debug" json:"debug"`
}

// Config 完整配置结构
type Config struct {
	Domain string `yaml:"domain" json:"domain"`
	APIs   []API  `yaml:"apis" json:"apis"`
	Server Server `yaml:"server" json:"server"`
}
