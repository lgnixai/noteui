package configs

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	RedisURL    string
	ServerPort  string
	Env         string // 新增环境变量标识
	CORSOrigin  string // 新增 CORS 配置
}

const (
	defaultPort = "8080" // 默认端口分离常量
	devEnv      = "development"
)

func LoadConfig() *Config {
	// 优先加载.env文件（开发环境）
	if err := godotenv.Load(); err == nil {
		log.Println("Using .env file configuration")
	} else {
		log.Printf("No .env file found: %v", err)
	}

	config := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		RedisURL:    os.Getenv("REDIS_URL"),
		ServerPort:  os.Getenv("SERVER_PORT"),
		Env:         os.Getenv("APP_ENV"),
		CORSOrigin:  os.Getenv("CORS_ORIGIN"),
	}

	// 设置默认环境
	if config.Env == "" {
		config.Env = devEnv
	}

	// 开发环境默认值
	if config.Env == devEnv {
		if config.DatabaseURL == "" {
			config.DatabaseURL = "postgres://user:password@localhost:5432/airtable_dev?sslmode=disable"
			log.Printf("DEV: Using default database URL: %s", config.DatabaseURL)
		}
		if config.RedisURL == "" {
			config.RedisURL = "redis://localhost:6379/0"
			log.Printf("DEV: Using default Redis URL: %s", config.RedisURL)
		}
		if config.CORSOrigin == "" {
			config.CORSOrigin = "http://localhost:3000"
			log.Printf("DEV: Using default CORS origin: %s", config.CORSOrigin)
		}
	}

	// 端口处理逻辑优化
	if config.ServerPort == "" {
		config.ServerPort = ":" + defaultPort
	} else if !strings.HasPrefix(config.ServerPort, ":") {
		config.ServerPort = ":" + config.ServerPort
	}

	return config
}
