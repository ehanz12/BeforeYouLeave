package configs

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// struct untuk config database
type Config struct {
	DBName               string
	DBUser               string
	DBPassword           string
	DBHost               string
	DBPort               string
	Port                 string
	DBMaxOpenConns       int
	DBMaxIdleConns       int
	DBConnMaxLifetimeMin int
	DBConnMaxIdleTimeMin int
}

// variable dari struct
var AppConfig *Config

func LoadEnv() {
	// cek env kalo ga ada kasih error
	if err := godotenv.Load(); err != nil {
		log.Println("Error Not Found file .env !⚠️")
	}

	// instalasi untuk config
	AppConfig = &Config{
		Port:                 os.Getenv("PORT"),
		DBName:               os.Getenv("DB_NAME"),
		DBPassword:           os.Getenv("DB_PASSWORD"),
		DBUser:               os.Getenv("DB_USER"),
		DBHost:               os.Getenv("DB_HOST"),
		DBPort:               os.Getenv("DB_PORT"),
		DBMaxOpenConns:       getEnvInt("DB_MAX_OPEN_CONNS", 5),
		DBMaxIdleConns:       getEnvInt("DB_MAX_IDLE_CONNS", 3),
		DBConnMaxLifetimeMin: getEnvInt("DB_CONN_MAX_LIFETIME_MIN", 5),
		DBConnMaxIdleTimeMin: getEnvInt("DB_CONN_MAX_IDLE_TIME_MIN", 3),
	}
}

func getEnvInt(key string, fallback int) int {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(val)
	if err != nil || parsed <= 0 {
		log.Printf("Warning: %s tidak valid (%q), pakai default %d", key, val, fallback)
		return fallback
	}
	return parsed
}

func getEnvFloat(key string, fallback float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(val, 64)
	if err != nil {
		log.Printf("Warning: %s tidak valid (%q), pakai default %v", key, val, fallback)
		return fallback
	}
	return parsed
}