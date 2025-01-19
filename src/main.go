package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/PayCryps/WatchdogGo/src/db"
	"github.com/PayCryps/WatchdogGo/src/monitor/docker"
	"github.com/PayCryps/WatchdogGo/src/server"
	"github.com/PayCryps/WatchdogGo/src/utils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
)

func monitorRoutine(logger zerolog.Logger, stop chan struct{}) {
	logger.Info().Msg("Monitor thread started")

	HaltTime := os.Getenv("HALT_TIME")
	if HaltTime == "" {
		HaltTime = "10"
	}
	haltDuration, err := strconv.Atoi(HaltTime)
	if err != nil {
		fmt.Println("Error converting HALT_TIME to an integer:", err)
		return
	}
	ticker := time.NewTicker(time.Second * time.Duration(haltDuration))
	defer ticker.Stop()

	dockerCli := docker.CreateDockerClient()

	for {
		select {
		case <-ticker.C:
			// Todo: Get all the containers required to monitor and check if they are running
			desiredConfig := container.Config{Image: "postgres:15.0-alpine"}
			desiredName := "/postgres"

			containers := []docker.ContainerDetails{}
			containers = append(containers, docker.ContainerDetails{Name: desiredName, Configs: desiredConfig})

			statusList := docker.IsContainerRunning(dockerCli, containers, logger)

			for _, status := range statusList {
				if !status.IsRunning {
					if status.ContainerID != "" {
						logger.Info().Msg(fmt.Sprintf("Restarting %s container", status.Name))
						docker.RestartContainer(dockerCli, status.ContainerID, logger)
					} else {
						// fetch the host configs from db based on the container name
						network := "watchdog"
						volume := mount.Mount{
							Type:   mount.TypeVolume,
							Source: "data",
							Target: "/var/lib/postgresql/data",
						}

						hostConfig := container.HostConfig{
							Mounts:      []mount.Mount{volume},
							NetworkMode: container.NetworkMode(network),
						}

						docker.CreateAndStartContainer(dockerCli, desiredConfig, hostConfig, desiredName, logger)
					}
				}
			}

		case <-stop:
			logger.Info().Msg("Monitor thread exiting")
			return
		}
	}
}

func startServer(logger zerolog.Logger) {
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

func main() {
	logger := utils.SetupLogger()

	err := godotenv.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load .env file")
	}

	db.InitDB(logger)
	defer db.CloseDB(logger)

	stop := make(chan struct{})

	go monitorRoutine(logger, stop)
	startServer(logger)

	close(stop)
}
