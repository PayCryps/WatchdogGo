package docker

import "github.com/docker/docker/api/types/container"

type ContainerDetails struct {
	Name    string
	Configs container.Config
}

type ContainerStatus struct {
	IsRunning   bool
	ContainerID string
	Name        string
}
