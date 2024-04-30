package queue

import (
	"context"
	"log"

	"github.com/SwarnimWalavalkar/container_provisioning_engine/database"
	"github.com/SwarnimWalavalkar/container_provisioning_engine/services"
	"github.com/SwarnimWalavalkar/container_provisioning_engine/types"
)

type CreateDeploymentTask struct {
	Db                   *database.Database
	Docker               *services.DockerService
	DeploymentAttributes *types.Deployment
	ImageTag             string
	Subdomain            string
	EnvArray             []string
	ContainerPort        int
	AuthString           string
}

func (task CreateDeploymentTask) Process() error {
	log.Println("ADDED DEPLOYMENT CREATE TASK TO QUEUE")
	log.Printf("%+v\n", task)

	containerId, err := task.Docker.ProvisionContainer(context.Background(), task.ImageTag, task.Subdomain, task.EnvArray, task.ContainerPort, task.AuthString)
	if err != nil {
		log.Printf("error provisioning container: %s\n", err.Error())
		return err
	}

	if _, err := task.Db.UpdateDeployment(context.Background(), types.DeploymentAttributes{UUID: *task.DeploymentAttributes.UUID, ImageTag: *task.DeploymentAttributes.ImageTag, Subdomain: *task.DeploymentAttributes.Subdomain, Port: task.DeploymentAttributes.Port, ContainerId: &containerId, Status: "READY"}); err != nil {
		log.Printf("error updating deployment row: %s\n", err.Error())
		return err
	}

	return nil
}
