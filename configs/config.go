package configs

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const defaultTimezone = "Asia/Jakarta"

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

	// Location = timezone aplikasi. Harus sama dengan loc= di DSN database.go,
	// kalau tidak tanggal (H-1 vs hari-H) akan bergeser.
	Location *time.Location
	// ReminderHorizonDays = berapa hari ke depan reminder di-generate.
	ReminderHorizonDays int
	// ReminderGenerateIntervalMin = seberapa sering worker membuat reminder baru.
	ReminderGenerateIntervalMin int
	// ReminderDispatchIntervalMin = seberapa sering pending -> sent.
	ReminderDispatchIntervalMin int
	GoogleClientID              string
	// JWTSecret = kunci penandatangan token. Wajib di-set, kalau kosong
	// login akan gagal saat membuat token.
	JWTSecret string
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
		Port:                        os.Getenv("PORT"),
		DBName:                      os.Getenv("DB_NAME"),
		DBPassword:                  os.Getenv("DB_PASSWORD"),
		DBUser:                      os.Getenv("DB_USER"),
		DBHost:                      os.Getenv("DB_HOST"),
		DBPort:                      os.Getenv("DB_PORT"),
		DBMaxOpenConns:              getEnvInt("DB_MAX_OPEN_CONNS", 5),
		DBMaxIdleConns:              getEnvInt("DB_MAX_IDLE_CONNS", 3),
		DBConnMaxLifetimeMin:        getEnvInt("DB_CONN_MAX_LIFETIME_MIN", 5),
		DBConnMaxIdleTimeMin:        getEnvInt("DB_CONN_MAX_IDLE_TIME_MIN", 3),
		Location:                    loadLocation(os.Getenv("APP_TIMEZONE")),
		ReminderHorizonDays:         getEnvInt("REMINDER_HORIZON_DAYS", 14),
		ReminderGenerateIntervalMin: getEnvInt("REMINDER_GENERATE_INTERVAL_MIN", 60),
		ReminderDispatchIntervalMin: getEnvInt("REMINDER_DISPATCH_INTERVAL_MIN", 5),
		GoogleClientID:              os.Getenv("GOOGLE_CLIENT_ID"),
		JWTSecret:                   os.Getenv("JWT_SECRET"),
	}

	// Gagal cepat: tanpa JWT_SECRET, register tetap jalan tapi login selalu
	// gagal di tengah proses. Lebih baik server menolak start sejak awal.
	if AppConfig.JWTSecret == "" {
		log.Fatal("⚠️ JWT_SECRET wajib di-set (lihat .env.example)")
	}
}

// loadLocation = timezone dari env, fallback Asia/Jakarta.
func loadLocation(name string) *time.Location {
	if name == "" {
		name = defaultTimezone
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		log.Printf("Warning: APP_TIMEZONE %q tidak valid, pakai default %s", name, defaultTimezone)
		loc, fallbackErr := time.LoadLocation(defaultTimezone)
		if fallbackErr != nil {
			return time.FixedZone(defaultTimezone, 7*60*60)
		}
		return loc
	}
	return loc
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
