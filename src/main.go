package main

import (
	"os"

	"github.com/PayCryps/WatchdogGo/src/db"
	"github.com/PayCryps/WatchdogGo/src/server"
	"github.com/PayCryps/WatchdogGo/src/utils"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	logger := utils.SetupLogger()

	err := godotenv.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load .env file")
	}

	db.InitDB(logger)
	defer db.CloseDB(logger)

	r := gin.Default()
	server.RegisterRoutes(r, logger)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info().Msg("Server starting at port: " + port)

	if err := r.Run(":" + port); err != nil {
		logger.Fatal().Err(err).Msg("Server failed to start")
	}
}
