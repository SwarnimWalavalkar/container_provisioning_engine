package handlers

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/SwarnimWalavalkar/aether/database"
	"github.com/SwarnimWalavalkar/aether/services"
	"github.com/SwarnimWalavalkar/aether/types"
	"github.com/SwarnimWalavalkar/aether/utils"
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

func CreateDeployment(db *database.Database, docker *services.DockerService) gin.HandlerFunc {
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

		containerId, err := docker.ProvisionContainer(c, deploymentReq.ImageTag, deploymentReq.Subdomain, envArray, containerPort, authString)

		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		deployment, err := db.CreateDeployment(c.Request.Context(), types.DeploymentAttributes{UserUUID: userUUID.(string), Subdomain: deploymentReq.Subdomain, ImageTag: deploymentReq.ImageTag, ContainerId: containerId, Port: containerPort})
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, deployment)
	}
}

func UpdateDeployment(db *database.Database) gin.HandlerFunc {
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

		deployment, err := db.UpdateDeployment(c.Request.Context(), types.DeploymentAttributes{UUID: uuid, Subdomain: updateDeploymentReq.Subdomain, ImageTag: updateDeploymentReq.ImageTag, ContainerId: fmt.Sprintf("%x", rand.Uint64())})
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, deployment)
	}
}

func DeleteDeployment(db *database.Database, docker *services.DockerService) gin.HandlerFunc {
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
			c.Error(err)
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		if *existingDeployment.UserId != *user.ID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		if err := docker.RemoveContainer(c, *existingDeployment.ContainerId); err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := db.DeleteDeployment(c.Request.Context(), uuid); err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusOK)
	}
}
