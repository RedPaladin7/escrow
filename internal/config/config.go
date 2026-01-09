package config

import (
	"os"
	"strconv"
)

type Config struct {
	Version       string
	WSPort        string
	APIPort       string
	MaxPlayers    int
	EnableHTTPS   bool
	InitialPeer   string
	ReadTimeout   int
	WriteTimeout  int
	PingInterval  int
}

func (c *Config) GetWSAddr() string {
	return ":" + c.WSPort
}

func (c *Config) GetAPIAddr() string {
	return ":" + c.APIPort
}

func LoadFromEnv() *Config {
	cfg := &Config{
		Version:      getEnv("POKER_VERSION", "2.0.0"),
		WSPort:       getEnv("WS_PORT", "3000"),
		APIPort:      getEnv("API_PORT", "8080"),
		MaxPlayers:   getEnvInt("MAX_PLAYERS", 6),
		EnableHTTPS:  getEnvBool("ENABLE_HTTPS", false),
		InitialPeer:  getEnv("INITIAL_PEER", ""),
		ReadTimeout:  getEnvInt("READ_TIMEOUT", 60),
		WriteTimeout: getEnvInt("WRITE_TIMEOUT", 10),
		PingInterval: getEnvInt("PING_INTERVAL", 30),
	}
	return cfg
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			return boolVal
		}
	}
	return defaultVal
}
