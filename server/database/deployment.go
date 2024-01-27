package database

import (
	"context"

	"github.com/SwarnimWalavalkar/aether/types"
)

func (d *Database) GetDeployment(ctx context.Context, uuidOrSubdomain string) (types.Deployment, error) {
	var deployment types.Deployment
	query := `SELECT * FROM deployments WHERE uuid = $1 OR sub_domain = $1`

	if err := d.Client.GetContext(ctx, &deployment, query, uuidOrSubdomain); err != nil {
		return types.Deployment{}, err
	}

	return deployment, nil
}

// @TODO: Paginate this
func (d *Database) GetAllDeploymentsForUser(ctx context.Context, userUUID string) ([]types.Deployment, error) {
	var deployments []types.Deployment
	query := `SELECT * FROM deployments WHERE user_id = (SELECT id FROM users WHERE uuid = $1)`

	if err := d.Client.SelectContext(ctx, &deployments, query, userUUID); err != nil {
		return []types.Deployment{}, err
	}

	if deployments == nil {
		return []types.Deployment{}, nil
	}

	return deployments, nil
}

func (d *Database) CreateDeployment(ctx context.Context, deploymentAttributes types.DeploymentAttributes) (types.Deployment, error) {
	user, err := d.GetUserByUUID(ctx, deploymentAttributes.UserUUID)
	if err != nil {
		return types.Deployment{}, err
	}

	deployment := types.Deployment{
		UserId:       user.ID,
		Subdomain:    &deploymentAttributes.Subdomain,
		ImageTag:     &deploymentAttributes.ImageTag,
		ContainerId:  &deploymentAttributes.ContainerId,
		InternalPort: &deploymentAttributes.InternalPort,
	}

	if _, err := d.Client.NamedExecContext(ctx, `INSERT INTO deployments (user_id, sub_domain, image_tag, container_id, internal_port) VALUES (:user_id, :sub_domain, :image_tag, :container_id, :internal_port)`, deployment); err != nil {
		return types.Deployment{}, err
	}

	createdDeployment, err := d.GetDeployment(ctx, deploymentAttributes.Subdomain)
	if err != nil {
		return types.Deployment{}, err
	}

	return createdDeployment, nil
}

func (d *Database) UpdateDeployment(ctx context.Context, deploymentAttributes types.DeploymentAttributes) (types.Deployment, error) {
	if _, err := d.Client.NamedExecContext(ctx, `UPDATE deployments SET image_tag = :image_tag, sub_domain = :sub_domain, container_id = :container_id WHERE uuid = :uuid`, deploymentAttributes); err != nil {
		return types.Deployment{}, err
	}

	deployment, err := d.GetDeployment(ctx, deploymentAttributes.UUID)
	if err != nil {
		return types.Deployment{}, err
	}

	return deployment, nil

}

func (d *Database) DeleteDeployment(ctx context.Context, uuid string) error {
	if _, err := d.Client.ExecContext(ctx, `DELETE FROM deployments WHERE uuid = $1`, uuid); err != nil {
		return err
	}

	return nil
}
