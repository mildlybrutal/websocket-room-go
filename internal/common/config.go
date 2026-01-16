package common

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Websocket  WSConfig         `mapstructure:"websocket"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
	Security   SecurityConfig   `mapstructure:"security"`
}

type ServerConfig struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstruture:"port"`
	ReadTimeout  string `mapstructure:"read_timeout"`
	WriteTimeout string `mapstructure:"write_timeout"`
	IdleTimeout  string `mapstructure:"idle_timeout"`
}

type WSConfig struct {
	MaxConnections    int           `mapstructure:"max_connections"`
	ReadBufferSize    int           `mapstructure:"read_buffer_size"`
	WriteBufferSize   int           `mapstructure:"write_buffer_size"`
	HandshakeTimeout  time.Duration `mapstructure:"handshake_timeout"`
	PongWait          time.Duration `mapstructure:"pong_wait"`
	PingPeriod        time.Duration `mapstructure:"ping_period"`
	WriteWait         time.Duration `mapstructure:"time_duration"`
	MaxMessageSize    int64         `mapstructure:"max_message_size"`
	EnableCompression bool          `mapstructure:"enable_compression"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type MonitoringConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	MetricsPath    string `mapstructure:"metrics_path"`
	PrometheusPort int    `mapstructure:"prometheus_port"`
}

type SecurityConfig struct {
	EnableTLS      bool     `mapstructure:"enable_tls"`
	CertFile       string   `mapstructure:"cert_file"`
	KeyFile        string   `mapstructure:"key_file"`
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	RequireAuth    bool     `mapstructure:"require_auth"`
	JWTSecret      string   `mapstructure:"jwt_secret"`
	RateLimitRPS   int      `mapstructure:"rate_limit_rps"`
	RateLimitBurst int      `mapstructure:"rate_limit_burst"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	SetDefaults()

	err := viper.ReadInConfig()

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			//Config file not found, use defaults
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	return &config, nil
}

func SetDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "15s")
	viper.SetDefault("server.write_timeout", "15s")
	viper.SetDefault("server.idle_timeout", "60s")

	// WebSocket defaults
	viper.SetDefault("websocket.max_connections", 10000)
	viper.SetDefault("websocket.read_buffer_size", 1024)
	viper.SetDefault("websocket.write_buffer_size", 1024)
	viper.SetDefault("websocket.handshake_timeout", "10s")
	viper.SetDefault("websocket.pong_wait", "60s")
	viper.SetDefault("websocket.ping_period", "54s")
	viper.SetDefault("websocket.write_wait", "10s")
	viper.SetDefault("websocket.max_message_size", 512*1024) // 512KB
	viper.SetDefault("websocket.enable_compression", true)

	// Security defaults
	viper.SetDefault("security.enable_tls", false)
	viper.SetDefault("security.allowed_origins", []string{"*"})
	viper.SetDefault("security.require_auth", false)
	viper.SetDefault("security.rate_limit_rps", 100)
	viper.SetDefault("security.rate_limit_burst", 200)

	// Monitoring defaults
	viper.SetDefault("monitoring.enabled", true)
	viper.SetDefault("monitoring.metrics_path", "/metrics")
	viper.SetDefault("monitoring.prometheus_port", 9090)
}
