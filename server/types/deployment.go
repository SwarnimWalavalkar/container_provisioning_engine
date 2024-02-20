package types

import "time"

type Deployment struct {
	ID     *int    `db:"id" json:"-"`
	UserId *int    `db:"user_id" json:"-"`
	UUID   *string `db:"uuid" json:"uuid"`

	ImageTag  *string `db:"image_tag" json:"imageTag"`
	Subdomain *string `db:"sub_domain" json:"subDomain"`

	ContainerId  *string `db:"container_id" json:"containerId"`
	InternalPort *int    `db:"internal_port" json:"-"`

	CreatedAt *time.Time `db:"created_at" json:"-"`
	UpdatedAt *time.Time `db:"updated_at" json:"-"`
}

type DeploymentAttributes struct {
	UUID     string `db:"uuid" json:"uuid"`
	UserUUID string `json:"userUUID"`

	Subdomain string `json:"subdomain" db:"sub_domain"`
	ImageTag  string `json:"imageTag" db:"image_tag"`

	ContainerId  string `json:"containerId" db:"container_id"`
	InternalPort int    `db:"internal_port" json:"-"`
}

type dockerAuth struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}
type CreateDeploymentRequest struct {
	Subdomain  string             `json:"subdomain" binding:"required"`
	ImageTag   string             `json:"imageTag" binding:"required"`
	EnvConfig  map[string]string `json:"envConfig"`
	DockerAuth *dockerAuth        `json:"dockerAuth"`
}

type UpdateDeploymentRequest struct {
	Subdomain string   `json:"subdomain"`
	ImageTag  string   `json:"imageTag"`
	EnvConfig []string `json:"envConfig"`
}
