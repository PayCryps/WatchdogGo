package db

import (
	"os"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/PayCryps/WatchdogGo/src/utils"
	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(logger zerolog.Logger) {
	var err error

	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		logger.Fatal().Msg("DATABASE_URL is not set")
		panic("DATABASE_URL is not set")
	}

	logger.Info().Msg("Connecting to the database")
	dsn := utils.GetDsn()
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to connect to the database")
	}

	logger.Info().Msg("Applying migrations")
	DB.AutoMigrate(&User{})
}

func CloseDB(logger zerolog.Logger) {
	db, err := DB.DB()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to get sql.DB from GORM")
	}

	if err := db.Close(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to close the database")
	}
}
