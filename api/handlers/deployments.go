package handlers

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/SwarnimWalavalkar/container_provisioning_engine/database"
	"github.com/SwarnimWalavalkar/container_provisioning_engine/queue"
	"github.com/SwarnimWalavalkar/container_provisioning_engine/services"
	"github.com/SwarnimWalavalkar/container_provisioning_engine/types"
	"github.com/SwarnimWalavalkar/container_provisioning_engine/utils"
	"github.com/docker/docker/api/types/registry"
	"github.com/gin-gonic/gin"
)

func GetDeployment(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := c.Param("uuid")

		deployment, err := db.GetDeployment(c.Request.Context(), uuid)

		if err != nil {
			c.Error(err)
			switch {
			case err == sql.ErrNoRows:
				c.JSON(http.StatusNotFound, map[string]interface{}{"error": fmt.Sprintf("Invalid UUID: %s", uuid)})
			default:
				c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Something went wrong"})
			}
			return
		}

		c.JSON(http.StatusOK, deployment)
	}
}

func GetAllDeploymentsForUser(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		userUUID, doesUserUUIDExists := c.Get("userUUID")

		if !doesUserUUIDExists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		if _, err := db.GetUserByUUID(c.Request.Context(), userUUID.(string)); err != nil {
			c.Error(err)
			switch {
			case err == sql.ErrNoRows:
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			default:
				c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Something went wrong"})
			}
			return
		}

		deployments, err := db.GetAllDeploymentsForUser(c.Request.Context(), userUUID.(string))
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Something went wrong"})
			return
		}

		c.JSON(http.StatusOK, deployments)
	}
}

func CreateDeployment(db *database.Database, docker *services.DockerService, taskDispatcher *queue.TaskDispatcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		var deploymentReq types.CreateDeploymentRequest
		if err := c.ShouldBindJSON(&deploymentReq); err != nil {
			c.Error(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userUUID, ok := c.Get("userUUID")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		if _, err := db.GetDeployment(c.Request.Context(), deploymentReq.Subdomain); err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Subdomain already exists"})
			return
		}

		if len(deploymentReq.EnvConfig) == 0 {
			deploymentReq.EnvConfig = make(map[string]string)
		}

		var containerPort int
		providedPortStr, exists := deploymentReq.EnvConfig["PORT"]
		if exists {
			providedPort, err := strconv.Atoi(providedPortStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "The value of port must be a valid port number"})
				return
			}
			containerPort = providedPort
		} else {
			availablePort, err := utils.GetFreePort()
			if err != nil {
				c.Error(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			containerPort = availablePort

			deploymentReq.EnvConfig["PORT"] = strconv.Itoa(availablePort)
		}

		var envArray []string
		for key, value := range deploymentReq.EnvConfig {
			envArray = append(envArray, fmt.Sprintf("%s=%s", key, value))
		}

		authString := ""
		if deploymentReq.DockerAuth != nil {
			authConfig := registry.AuthConfig{Username: deploymentReq.DockerAuth.Username, Password: deploymentReq.DockerAuth.Password}

			encodedJSON, err := json.Marshal(authConfig)
			if err != nil {
				panic(err)
			}

			authString = base64.URLEncoding.EncodeToString(encodedJSON)
		}

		deployment, err := db.CreateDeployment(c.Request.Context(), types.DeploymentAttributes{UserUUID: userUUID.(string), Subdomain: deploymentReq.Subdomain, ImageTag: deploymentReq.ImageTag, ContainerId: nil, Status: "PENDING", Port: &containerPort})
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		taskDispatcher.Enqueue(queue.CreateDeploymentTask{Db: db, Docker: docker, DeploymentAttributes: &deployment, ImageTag: *deployment.ImageTag, Subdomain: *deployment.Subdomain, EnvArray: envArray, ContainerPort: containerPort, AuthString: authString})

		c.JSON(http.StatusOK, gin.H{"uuid": deployment.UUID, "status": deployment.Status})
	}
}

func UpdateDeployment(db *database.Database, docker *services.DockerService, taskDispatcher *queue.TaskDispatcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := c.Param("uuid")
		var updateDeploymentReq types.UpdateDeploymentRequest
		if err := c.ShouldBindJSON(&updateDeploymentReq); err != nil {
			c.Error(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userUUID, doesUserUUIDExists := c.Get("userUUID")

		if !doesUserUUIDExists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		existingDeployment, err := db.GetDeployment(c.Request.Context(), uuid)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid deployment uuid"})
			return
		}

		user, err := db.GetUserByUUID(c.Request.Context(), userUUID.(string))
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		if *existingDeployment.UserId != *user.ID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		existingContainerEnv, err := docker.GetContainerEnv(c, *existingDeployment.ContainerId)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var envMap map[string]string = existingContainerEnv
		var imageTag string = *existingDeployment.ImageTag
		var subdomain string = *existingDeployment.Subdomain
		var containerPort int = *existingDeployment.Port

		if len(*updateDeploymentReq.EnvConfig) != 0 {
			envMap = utils.MergeMaps(existingContainerEnv, *updateDeploymentReq.EnvConfig)

			providedPortStr, exists := (*updateDeploymentReq.EnvConfig)["PORT"]
			if exists {
				providedPort, err := strconv.Atoi(providedPortStr)
				if err != nil {
					c.Error(err)
					c.JSON(http.StatusBadRequest, gin.H{"error": "The value of port must be a valid port number"})
					return
				}
				containerPort = providedPort
			}
		}

		var envArray []string
		for key, value := range envMap {
			envArray = append(envArray, fmt.Sprintf("%s=%s", key, value))
		}

		if updateDeploymentReq.ImageTag != nil {
			imageTag = *updateDeploymentReq.ImageTag
		}

		if updateDeploymentReq.Subdomain != nil && *updateDeploymentReq.Subdomain != *existingDeployment.Subdomain {
			if _, err := db.GetDeployment(c.Request.Context(), *updateDeploymentReq.Subdomain); err == nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Subdomain already exists"})
				return
			} else {
				subdomain = *updateDeploymentReq.Subdomain
			}
		}

		authString := ""
		if updateDeploymentReq.DockerAuth != nil {
			authConfig := registry.AuthConfig{Username: updateDeploymentReq.DockerAuth.Username, Password: updateDeploymentReq.DockerAuth.Password}

			encodedJSON, err := json.Marshal(authConfig)
			if err != nil {
				panic(err)
			}

			authString = base64.URLEncoding.EncodeToString(encodedJSON)
		}

		taskDispatcher.Enqueue(queue.UpdateDeploymentTask{Db: db, Docker: docker, DeploymentAttributes: &existingDeployment, ImageTag: imageTag, Subdomain: subdomain, EnvArray: envArray, ContainerPort: containerPort, AuthString: authString})

		c.JSON(http.StatusOK, gin.H{"uuid": existingDeployment.UUID, "update_queued": true})
	}
}

func DeleteDeployment(db *database.Database, docker *services.DockerService, taskDispatcher *queue.TaskDispatcher) gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := c.Param("uuid")

		userUUID, doesUserUUIDExists := c.Get("userUUID")

		if !doesUserUUIDExists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		existingDeployment, err := db.GetDeployment(c.Request.Context(), uuid)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Invalid deployment uuid"})
			return
		}

		user, err := db.GetUserByUUID(c.Request.Context(), userUUID.(string))
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		if *existingDeployment.UserId != *user.ID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		if *existingDeployment.Status == "DELETING" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Deployment is already being deleted"})
			return
		}

		if _, err := db.UpdateDeployment(context.Background(), types.DeploymentAttributes{UUID: *existingDeployment.UUID, ImageTag: *existingDeployment.ImageTag, Subdomain: *existingDeployment.Subdomain, Port: existingDeployment.Port, ContainerId: existingDeployment.ContainerId, Status: "DELETING"}); err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		taskDispatcher.Enqueue(queue.DeleteDeploymentTask{Db: db, Docker: docker, DeploymentAttributes: &existingDeployment})

		c.Status(http.StatusOK)
	}
}
