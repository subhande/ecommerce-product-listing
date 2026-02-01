package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads variables from .env into the process environment.
func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; relying on existing environment variables")
	}

	// PRINT ENV VARIABLES FOR DEBUGGING PURPOSES
	log.Println("Environment variables loaded:")
	for _, key := range []string{"DATABASE_URL", "DATABASE_NAME"} {
		log.Printf("%s=%s\n", key, os.Getenv(key))
	}
}
