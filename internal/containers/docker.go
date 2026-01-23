package containers

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/rs/zerolog/log"
)

// Contains settings for deploying a container
type DeployConfig struct {
	ContainerName string
	ImageTag      string
	Port          int
	EnvVars       map[string]string
	Slug          string
	BaseDomain    string
}

// Creates and starts a container with Traefik labels
func Deploy(ctx context.Context, cfg *DeployConfig) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv,
		client.WithAPIVersionNegotiation())
	if err != nil {
		return "", fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer cli.Close()

	// Stop and remove existing container with same name
	if err := stopAndRemove(ctx, cli, cfg.ContainerName); err != nil {
		log.Warn().Err(err).Str("container", cfg.ContainerName).
			Msg("Failed to stop existing container (may not exist)")
	}

	// Pull the image
	reader, err := cli.ImagePull(ctx, cfg.ImageTag, image.PullOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to pull image: %w", err)
	}
	io.Copy(io.Discard, reader) // Drain the reader
	reader.Close()

	// Convert env vars to slice format
	envSlice := make([]string, 0, len(cfg.EnvVars))
	for k, v := range cfg.EnvVars {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
	}

	// Traefik labels for dynamic routing
	hostname := fmt.Sprintf("%s.%s", cfg.Slug, cfg.BaseDomain)
	labels := map[string]string{
		"traefik.enable": "true",
		// HTTP Router
		fmt.Sprintf("traefik.http.routers.%s.rule", cfg.Slug):        fmt.Sprintf("Host(`%s`)", hostname),
		fmt.Sprintf("traefik.http.routers.%s.entrypoints", cfg.Slug): "web",
		// HTTPS Router
		fmt.Sprintf("traefik.http.routers.%s-secure.rule", cfg.Slug):        fmt.Sprintf("Host(`%s`)", hostname),
		fmt.Sprintf("traefik.http.routers.%s-secure.entrypoints", cfg.Slug): "websecure",
		fmt.Sprintf("traefik.http.routers.%s-secure.tls", cfg.Slug):         "true",
		// Service port
		fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", cfg.Slug): fmt.Sprintf("%d", cfg.Port),
		// RCNbuild metadata
		"rcnbuild.managed": "true",
		"rcnbuild.slug":    cfg.Slug,
	}

	// Add Let's Encrypt certresolver if TLS enabled
	if os.Getenv("TLS_ENABLED") == "true" {
		labels[fmt.Sprintf("traefik.http.routers.%s-secure.tls.certresolver", cfg.Slug)] = "letsencrypt"
	}

	// Container configuration
	containerCfg := &container.Config{
		Image:  cfg.ImageTag,
		Env:    envSlice,
		Labels: labels,
		ExposedPorts: nat.PortSet{
			nat.Port(fmt.Sprintf("%d/tcp", cfg.Port)): struct{}{},
		},
	}

	// Host configuration
	hostCfg := &container.HostConfig{
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyUnlessStopped,
		},
		Resources: container.Resources{
			Memory:   512 * 1024 * 1024, // 512MB limit
			NanoCPUs: 500000000,         // 0.5 CPU
		},
	}

	// Network configuration - connect to rcnbuild-network for Traefik
	networkCfg := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"rcnbuild-network": {},
		},
	}

	// Create the container
	resp, err := cli.ContainerCreate(ctx, containerCfg, hostCfg, networkCfg, nil,
		cfg.ContainerName)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Start the container
	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	log.Info().
		Str("container_id", resp.ID[:12]).
		Str("name", cfg.ContainerName).
		Str("hostname", hostname).
		Msg("Container started successfully")

	return resp.ID, nil
}

// Stop stops a running container
func Stop(ctx context.Context, containerID string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv,
		client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	timeout := 30 // seconds
	return cli.ContainerStop(ctx, containerID, container.StopOptions{
		Timeout: &timeout,
	})
}

// Remove removes a container
func Remove(ctx context.Context, containerID string) error {
	cli, err := client.NewClientWithOpts(client.FromEnv,
		client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer cli.Close()

	return cli.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: true,
	})
}

// Stops and removes a container by name
func stopAndRemove(ctx context.Context, cli *client.Client, name string) error {
	// Find container by name
	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("name", name),
		),
	})
	if err != nil {
		return err
	}

	for _, c := range containers {
		// Check if this is an exact name match (Docker prefixes with /)
		for _, n := range c.Names {
			if strings.TrimPrefix(n, "/") == name {
				log.Info().Str("container_id", c.ID[:12]).Msg("Stopping existing container")
				timeout := 30
				cli.ContainerStop(ctx, c.ID, container.StopOptions{Timeout: &timeout})
				cli.ContainerRemove(ctx, c.ID, container.RemoveOptions{Force: true})
				break
			}
		}
	}

	return nil
}

// Returns the logs for a container
func GetContainerLogs(ctx context.Context, containerID string, tail int) (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv,
		client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}
	defer cli.Close()

	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       fmt.Sprintf("%d", tail),
	}

	reader, err := cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return "", err
	}
	defer reader.Close()

	logs, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(logs), nil
}
