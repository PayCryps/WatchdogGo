package utils

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func SetupLogger() zerolog.Logger {
	logger := zerolog.New(
		zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339},
	).Level(zerolog.TraceLevel).With().Timestamp().Caller().Logger()

	return logger
}

// Helper function to check if SSL is already in the connection string
func ContainsSSL(connString string) bool {
	return len(connString) > 0 && connString[len(connString)-9:] == "?sslmode=disable"
}

func GetDsn() string {
	host := os.Getenv("DATABASE_HOST")
	user := os.Getenv("DATABASE_USER")
	pass := os.Getenv("DATABASE_PASS")
	dbname := os.Getenv("DATABASE_DBNAME")
	port := os.Getenv("DATABASE_PORT")
	ssl := os.Getenv("DATABASE_SSL")

	dsn := "host=" + host + " user=" + user + " password=" + pass + " dbname=" + dbname + " port=" + port + " sslmode=" + ssl

	return dsn
}

func Contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
