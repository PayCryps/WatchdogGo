package docker

import (
	"context"
	"fmt"

	"github.com/PayCryps/WatchdogGo/src/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/rs/zerolog"
)

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
					logger.Error().Msg(fmt.Sprintf("Container %s is not running", containerDetails.Name))
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
