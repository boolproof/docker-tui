package main

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

type DockerClientWrapper struct {
	client *client.Client
}

func (dc DockerClientWrapper) GetClient() *client.Client {
	if dc.client == nil {
		client, err := client.NewClientWithOpts(client.FromEnv)
		if err != nil {
			panic(err)
		}

		dc.client = client
	}

	return dc.client
}

func (dc DockerClientWrapper) GetDeamonEventStreams() (<-chan events.Message, <-chan error) {
	return dc.GetClient().Events(context.Background(), types.EventsOptions{})
}

func (dc DockerClientWrapper) GetContainerList(all bool) []types.Container {
	containers, err := dc.GetClient().ContainerList(context.Background(), types.ContainerListOptions{All: all})
	if err != nil {
		panic(err)
	}

	return containers
}

func (dc DockerClientWrapper) StartContainer(containerID string) {
	err := dc.GetClient().ContainerStart(context.Background(), containerID, types.ContainerStartOptions{})

	if err != nil {
		panic(err)
	}
}

func (dc DockerClientWrapper) StopContainer(containerID string) {
	err := dc.GetClient().ContainerStop(context.Background(), containerID, nil)

	if err != nil {
		panic(err)
	}
}
