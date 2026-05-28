package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	NodeEnv     string
	GatewayPort string
	UserPort    string
}

var (
	EnvCfg Env
)

func loadEnv() {
	if err := godotenv.Load(".env"); err != nil {
		godotenv.Load("../../.env")
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func init() {
	loadEnv()
	EnvCfg = Env{
		NodeEnv:     getEnv("NODE_ENV", "development"),
		GatewayPort: getEnv("GATEWAY_PORT", "8080"),
		UserPort:    getEnv("USER_PORT", "50051"),
	}
}
