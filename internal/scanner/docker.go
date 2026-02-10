package scanner

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DockerScanner struct {
	client *client.Client
}

func NewDockerScanner() (*DockerScanner, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	return &DockerScanner{client: cli}, nil
}

func (ds *DockerScanner) ListContainers(ctx context.Context) ([]ContainerInfo, error) {
	var found []ContainerInfo

	containers, err := ds.client.ContainerList(ctx, container.ListOptions{
		All: true,
	})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	for _, c := range containers {
		imageName := c.Image
		if len(c.Names) == 0 {
			continue
		}

		name := c.Names[0]
		if len(name) > 0 && name[0] == '/' {
			name = name[1:]
		}

		found = append(found, ContainerInfo{
			ID:    c.ID[:12],
			Name:  name,
			Image: imageName,
			State: c.State,
		})
	}

	return found, nil
}

func (ds *DockerScanner) Close() error {
	if ds.client != nil {
		return ds.client.Close()
	}
	return nil
}

type ContainerInfo struct {
	ID    string
	Name  string
	Image string
	State string
}
