package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
)

const (
	NETWORK_NAME              string = "traefik_default"
	TRAEFIK_ENTRYPOINT_NAME   string = "websecure"
	TRAEFIK_CERTRESOLVER_NAME string = "tlsresolver"
)

type DockerServiceType struct{}

func (d *DockerServiceType) ProvisionDockerContainer(ctx context.Context, image string, serviceName string, port string) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	_, err = cli.Ping(ctx)
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully initialized Docker client")

	authConfig := registry.AuthConfig{
		Username: os.Getenv("DOCKER_USERNAME"),
		Password: os.Getenv("DOCKER_PASSWORD"),
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		panic(err)
	}

	authStr := base64.URLEncoding.EncodeToString(encodedJSON)

	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{RegistryAuth: authStr})
	if err != nil {
		panic(err)
	}

	defer reader.Close()

	io.Copy(os.Stdout, reader)

	serviceHostname := fmt.Sprintf("%s.docker.localhost", serviceName)

	traefikLabels := map[string]string{
		"traefik.enable": "true",
		fmt.Sprintf("traefik.http.routers.%s.entrypoints", serviceName):               TRAEFIK_ENTRYPOINT_NAME,
		fmt.Sprintf("traefik.http.routers.%s.tls.certresolver", serviceName):          TRAEFIK_CERTRESOLVER_NAME,
		fmt.Sprintf("traefik.http.routers.%s.rule", serviceName):                      fmt.Sprintf("Host(`%s`)", serviceHostname),
		fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", serviceName): port,
	}

	cont, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image:    image,
			Labels:   traefikLabels,
			Hostname: serviceHostname,
			Env:      []string{fmt.Sprintf("PORT=%s", port)},
		},
		&container.HostConfig{}, &network.NetworkingConfig{EndpointsConfig: map[string]*network.EndpointSettings{NETWORK_NAME: {NetworkID: NETWORK_NAME}}}, nil, serviceName)
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}
	fmt.Printf("Container ID %s: started\n", cont.ID)
	fmt.Printf("Service running on: https://%s\n", serviceHostname)
}
