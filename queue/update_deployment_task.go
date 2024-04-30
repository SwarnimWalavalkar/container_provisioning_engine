package queue

import (
	"context"
	"log"

	"github.com/SwarnimWalavalkar/container_provisioning_engine/database"
	"github.com/SwarnimWalavalkar/container_provisioning_engine/services"
	"github.com/SwarnimWalavalkar/container_provisioning_engine/types"
)

type UpdateDeploymentTask struct {
	Db                   *database.Database
	Docker               *services.DockerService
	DeploymentAttributes *types.Deployment
	ImageTag             string
	Subdomain            string
	EnvArray             []string
	ContainerPort        int
	AuthString           string
}

func (task UpdateDeploymentTask) Process() error {
	log.Println("ADDED DEPLOYMENT UPDATE TASK TO QUEUE")
	log.Printf("%+v\n", task)

	if err := task.Docker.RemoveContainer(context.Background(), *task.DeploymentAttributes.ContainerId); err != nil {
		log.Printf("error removing container: %s\n", err.Error())
		return err
	}

	containerId, err := task.Docker.ProvisionContainer(context.Background(), task.ImageTag, task.Subdomain, task.EnvArray, task.ContainerPort, task.AuthString)
	if err != nil {
		log.Printf("error provisioning container: %s\n", err.Error())
		return err
	}

	if _, err := task.Db.UpdateDeployment(context.Background(), types.DeploymentAttributes{UUID: *task.DeploymentAttributes.UUID, ImageTag: task.ImageTag, Subdomain: task.Subdomain, Port: &task.ContainerPort, ContainerId: &containerId, Status: "READY"}); err != nil {
		log.Printf("error updating deployment row: %s\n", err.Error())
		return err
	}

	return nil
}
