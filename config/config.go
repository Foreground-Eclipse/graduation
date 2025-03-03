package config

import (
	"fmt"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	Config struct {
		Server   ServerConfig
		Database DatabaseConfig
		JWT      JWTConfig
	}

	ServerConfig struct {
		Address string `env:"SERVER_ADDRESS"`
	}

	DatabaseConfig struct {
		Host     string `env:"DATABASE_HOST"`
		Port     string `env:"DATABASE_PORT"`
		Name     string `env:"DATABASE_NAME"`
		User     string `env:"DATABASE_USER"`
		Password string `env:"DATABASE_PASSWORD"`
		SSLMode  string `env:"DATABASE_SSLMODE"`
	}

	JWTConfig struct {
		JWTSecret              string `env:"JWT_SECRET"`
		JWTExpirationTimeHours int    `env:"JWT_EXPIRATION_TIME_HOURS"`
	}
)

func MustLoad(filename string) *Config {
	configPath := fmt.Sprintf("./config/%s.env", filename)
	fmt.Println(os.Getwd())
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file doesnt exists %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cant read config: %s", err)
	}

	return &cfg

}
