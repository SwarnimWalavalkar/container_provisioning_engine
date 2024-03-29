package services

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/SwarnimWalavalkar/container_provisioning_engine/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
)

const (
	NETWORK_NAME              string = "traefik_default"
	TRAEFIK_ENTRYPOINT_NAME   string = "websecure"
	TRAEFIK_CERTRESOLVER_NAME string = "tlsresolver"
)

type DockerService struct {
	client *client.Client
}

func NewDockerService() (*DockerService, error) {
	client, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return &DockerService{}, err
	}

	return &DockerService{
		client: client,
	}, nil
}

func (d *DockerService) Ping(ctx context.Context) error {
	_, err := d.client.Ping(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (d *DockerService) ProvisionContainer(ctx context.Context, image string, serviceName string, envConfig []string, port int, authSting string) (string, error) {
	reader, err := d.client.ImagePull(ctx, image, types.ImagePullOptions{RegistryAuth: authSting})
	if err != nil {
		panic(err)
	}

	defer reader.Close()

	io.Copy(os.Stdout, reader)

	serviceHostname := fmt.Sprintf("%s.%s", serviceName, config.DEFAULT_HOSTNAME)

	traefikLabels := map[string]string{
		"traefik.enable": "true",
		fmt.Sprintf("traefik.http.routers.%s.entrypoints", serviceName):               TRAEFIK_ENTRYPOINT_NAME,
		fmt.Sprintf("traefik.http.routers.%s.tls.certresolver", serviceName):          TRAEFIK_CERTRESOLVER_NAME,
		fmt.Sprintf("traefik.http.routers.%s.rule", serviceName):                      fmt.Sprintf("Host(`%s`)", serviceHostname),
		fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", serviceName): fmt.Sprintf("%d", port),
	}

	cont, err := d.client.ContainerCreate(
		ctx,
		&container.Config{
			Image:    image,
			Labels:   traefikLabels,
			Hostname: serviceHostname,
			Env:      envConfig,
		},
		&container.HostConfig{}, &network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{NETWORK_NAME: {NetworkID: NETWORK_NAME}}}, nil, serviceName)
	if err != nil {
		panic(err)
	}

	if err := d.client.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
	fmt.Printf("Container ID %s: started\n", cont.ID)
	fmt.Printf("Service running on: https://%s\n", serviceHostname)

	return cont.ID, nil
}

func (d *DockerService) RemoveContainer(ctx context.Context, containerID string) error {
	if err := d.client.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		return err
	}

	if err := d.client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}); err != nil {
		return err
	}

	return nil
}

func (d *DockerService) GetContainerEnv(ctx context.Context, containerId string) (map[string]string, error) {
	resp, err := d.client.ContainerInspect(ctx, containerId)
	if err != nil {
		return map[string]string{}, err
	}

	// Convert environment variables to a map
	envMap := make(map[string]string)
	for _, env := range resp.Config.Env {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	return envMap, nil
}
