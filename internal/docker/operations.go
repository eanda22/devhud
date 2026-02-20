package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type Client struct {
	cli *client.Client
}

// creates a Docker client for container operations.
func NewClient() (*Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("docker client: %w", err)
	}
	return &Client{cli: cli}, nil
}

// starts a stopped container.
func (c *Client) Start(containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return fmt.Errorf("start container: %w", err)
	}
	return nil
}

// stops a running container with timeout.
func (c *Client) Stop(containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	timeout := 10
	if err := c.cli.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("stop container: %w", err)
	}
	return nil
}

// restarts a container with timeout.
func (c *Client) Restart(containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	timeout := 10
	if err := c.cli.ContainerRestart(ctx, containerID, container.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("restart container: %w", err)
	}
	return nil
}

// removes a container.
func (c *Client) Remove(containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{}); err != nil {
		return fmt.Errorf("remove container: %w", err)
	}
	return nil
}

// retrieves the last N lines of logs from a container.
func (c *Client) GetLogs(containerID string, lines int) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       strconv.Itoa(lines),
	}

	reader, err := c.cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return nil, fmt.Errorf("fetch logs: %w", err)
	}
	defer reader.Close()

	var logs []string
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 8 {
			logs = append(logs, line[8:])
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, fmt.Errorf("read logs: %w", err)
	}

	return logs, nil
}

// closes the Docker client connection.
func (c *Client) Close() error {
	if c.cli != nil {
		return c.cli.Close()
	}
	return nil
}

// returns the underlying Docker client for direct API access.
func (c *Client) GetRawClient() *client.Client {
	return c.cli
}

type DiskUsage struct {
	Containers     int64
	Images         int64
	Total          int64
	ContainerSizes map[string]int64
}

// returns disk usage for containers and images.
func (c *Client) GetDiskUsage() (*DiskUsage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get container list with size information
	containers, err := c.cli.ContainerList(ctx, container.ListOptions{
		All:  true,
		Size: true,
	})
	if err != nil {
		return nil, fmt.Errorf("container list: %w", err)
	}

	var containerSize int64
	containerSizes := make(map[string]int64)
	for _, cnt := range containers {
		size := cnt.SizeRw
		if size == 0 {
			size = cnt.SizeRootFs
		}
		containerSize += size
		// Store both full ID and short ID (first 12 chars) for compatibility
		containerSizes[cnt.ID] = size
		if len(cnt.ID) >= 12 {
			containerSizes[cnt.ID[:12]] = size
		}
	}

	// Get image sizes from DiskUsage
	diskUsage, err := c.cli.DiskUsage(ctx, types.DiskUsageOptions{})
	if err != nil {
		return nil, fmt.Errorf("disk usage: %w", err)
	}

	var imageSize int64
	for _, image := range diskUsage.Images {
		imageSize += image.Size
	}

	return &DiskUsage{
		Containers:     containerSize,
		Images:         imageSize,
		Total:          containerSize + imageSize,
		ContainerSizes: containerSizes,
	}, nil
}

// returns JSON representation of container inspect data.
func (c *Client) GetInspect(containerID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	inspectData, err := c.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return "", fmt.Errorf("inspect failed: %w", err)
	}

	jsonBytes, err := json.MarshalIndent(inspectData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("json marshal: %w", err)
	}

	return string(jsonBytes), nil
}
