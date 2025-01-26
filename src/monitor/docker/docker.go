package docker

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/PayCryps/WatchdogGo/src/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog"
)

func MonitorDocker(logger zerolog.Logger, dockerStop chan struct{}) {
	logger.Info().Msg("Docker Monitor thread started")

	HaltTime := os.Getenv("DOCKER_HALT_TIME")
	if HaltTime == "" {
		HaltTime = "10"
	}
	haltDuration, err := strconv.Atoi(HaltTime)
	if err != nil {
		logger.Error().Msgf("Error converting PROCESS_HALT_TIME to an integer: %s", err)
		return
	}
	ticker := time.NewTicker(time.Second * time.Duration(haltDuration))
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			monitorDocker(logger)

		case <-dockerStop:
			logger.Info().Msg("Docker thread exiting")
			return
		}
	}
}

func monitorDocker(logger zerolog.Logger) {
	dockerCli := CreateDockerClient()

	// Todo: Get all the containers required to monitor and check if they are running
	desiredConfig := container.Config{Image: "postgres:15.0-alpine"}
	desiredName := "/postgres"

	containers := []ContainerDetails{}
	containers = append(containers, ContainerDetails{Name: desiredName, Configs: desiredConfig})

	statusList := IsContainerRunning(dockerCli, containers, logger)

	for _, status := range statusList {
		if !status.IsRunning {
			if status.ContainerID != "" {
				logger.Warn().Msgf("Restarting %s container", status.Name)
				RestartContainer(dockerCli, status.ContainerID, logger)
			} else {

				DockerStart := os.Getenv("DOCKER_START")
				if DockerStart == "FALSE" {
					logger.Info().Msg("DOCKER_START is set to false, not starting container")
					continue
				}

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

				CreateAndStartContainer(dockerCli, desiredConfig, hostConfig, desiredName, logger)
			}
		}
	}
}

func CreateDockerClient() *client.Client {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	return cli
}

func GetDockerContainers(client *client.Client, logger zerolog.Logger) []types.Container {
	containers, err := client.ContainerList(context.Background(), container.ListOptions{All: true})
	if err != nil {
		panic(err)
	}

	return containers
}

func IsContainerRunning(cli *client.Client, desiredContainers []ContainerDetails, logger zerolog.Logger) []ContainerStatus {
	containerList := GetDockerContainers(cli, logger)

	containerStatusList := []ContainerStatus{}

	for _, containerDetails := range desiredContainers {
		found := false
		for _, container := range containerList {
			if container.Image == containerDetails.Configs.Image && utils.Contains(container.Names, containerDetails.Name) {
				if container.State == "running" {
					containerStatusList = append(containerStatusList, ContainerStatus{IsRunning: true, ContainerID: container.ID, Name: containerDetails.Name})
				} else {
					logger.Error().Msgf("Container %s is not running", containerDetails.Name)
					containerStatusList = append(containerStatusList, ContainerStatus{IsRunning: false, ContainerID: container.ID, Name: containerDetails.Name})
				}
				found = true
				break
			}
		}

		if !found {
			logger.Error().Msg(fmt.Sprintf("Container %s not found in docker", containerDetails.Name))
			containerStatusList = append(containerStatusList, ContainerStatus{IsRunning: false, ContainerID: "", Name: containerDetails.Name})
		}
	}

	return containerStatusList
}

func CreateContainer(cli *client.Client, desiredConfig container.Config, desiredHostConfig container.HostConfig, desiredName string, logger zerolog.Logger) {
	_, err := cli.ContainerCreate(context.Background(), &desiredConfig, &desiredHostConfig, nil, nil, desiredName)
	if err != nil {
		logger.Error().Msg(fmt.Sprintf("Failed to create container %s", desiredName))
	}
}

func RemoveContainer(cli *client.Client, containerID string, logger zerolog.Logger) {
	err := cli.ContainerRemove(context.Background(), containerID, container.RemoveOptions{})
	if err != nil {
		logger.Error().Msg(fmt.Sprintf("Failed to stop container %s", containerID))
	}
}

func FindAndRemoveContainer(cli *client.Client, desiredConfig container.Config, desiredName string, logger zerolog.Logger) {
	containerList := GetDockerContainers(cli, logger)

	for _, container := range containerList {
		if container.Image == desiredConfig.Image && utils.Contains(container.Names, desiredName) {
			RemoveContainer(cli, container.ID, logger)
		}
	}
}

func CreateAndStartContainer(cli *client.Client, desiredConfig container.Config, hostConfig container.HostConfig, desiredContainerName string, logger zerolog.Logger) {
	ctx := context.Background()

	FindAndRemoveContainer(cli, desiredConfig, desiredContainerName, logger)

	containerResp, err := cli.ContainerCreate(ctx, &desiredConfig, &hostConfig, nil, nil, desiredContainerName)
	if err != nil {
		logger.Error().Msg(fmt.Sprintf("Failed to create container %s", desiredContainerName))
	}

	if err := cli.ContainerStart(ctx, containerResp.ID, container.StartOptions{}); err != nil {
		logger.Error().Msg(fmt.Sprintf("Failed to start container %s", desiredContainerName))
	}

	logger.Info().Msg(fmt.Sprintf("Container created and started: %s", containerResp.ID))
}

func RestartContainer(cli *client.Client, containerID string, logger zerolog.Logger) {
	ctx := context.Background()

	err := cli.ContainerRestart(ctx, containerID, container.StopOptions{})
	if err != nil {
		logger.Error().Msg(fmt.Sprintf("Failed to restart container %s", containerID))
	}
}
