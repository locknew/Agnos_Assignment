package config

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
}

func Load() Config {
	env, err := godotenv.Read()
	if err != nil {
		log.Fatal(".env file is required")
	}

	dbUser := env["DB_USER"]
	dbPass := env["DB_PASSWORD"]
	dbHost := env["DB_HOST"]
	dbPort := env["DB_PORT"]
	dbName := env["DB_NAME"]

	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "5432"
	}
	if dbUser == "" || dbName == "" {
		log.Fatal("DB_USER and DB_NAME are required in .env")
	}

	databaseURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName,
	)

	jwtSecret := env["JWT_SECRET"]
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET is required in .env")
	}

	return Config{
		DatabaseURL: databaseURL,
		JWTSecret:   jwtSecret,
	}
}
