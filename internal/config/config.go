package config

import "os"

type Config struct {
	ServerPort  string
	DatabaseURL string
}

func LoadConfig() Config {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "mongodb://localhost:27017"
	}

	return Config{
		ServerPort:  port,
		DatabaseURL: dbURL,
	}
}
