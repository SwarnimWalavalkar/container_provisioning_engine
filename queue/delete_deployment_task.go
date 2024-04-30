package queue

import (
	"context"
	"log"

	"github.com/SwarnimWalavalkar/container_provisioning_engine/database"
	"github.com/SwarnimWalavalkar/container_provisioning_engine/services"
	"github.com/SwarnimWalavalkar/container_provisioning_engine/types"
)

type DeleteDeploymentTask struct {
	Db                   *database.Database
	Docker               *services.DockerService
	DeploymentAttributes *types.Deployment
}

func (task DeleteDeploymentTask) Process() error {
	log.Println("ADDED DEPLOYMENT DELETE TASK TO QUEUE")
	log.Printf("%+v\n", task)

	if err := task.Docker.RemoveContainer(context.Background(), *task.DeploymentAttributes.ContainerId); err != nil {
		log.Printf("error removing container: %s\n", err.Error())
		return err
	}

	if err := task.Db.DeleteDeployment(context.Background(), *task.DeploymentAttributes.UUID); err != nil {
		log.Printf("error deleting deployment row: %s\n", err.Error())
		return err
	}

	return nil
}
