package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/SwarnimWalavalkar/aether/config"
	"github.com/SwarnimWalavalkar/aether/database"
	"github.com/SwarnimWalavalkar/aether/types"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func GetUser(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		uuid := c.Param("uuid")

		user, err := db.GetUserByUUID(c.Request.Context(), uuid)

		if err != nil {
			c.Error(err)
			switch {
			case err == sql.ErrNoRows:
				c.JSON(http.StatusBadRequest, map[string]interface{}{"error": fmt.Sprintf("Invalid UUID: %s", uuid)})
			default:
				c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Something went wrong"})
			}
			return
		}

		c.JSON(http.StatusOK, user)
	}
}
func CreateUser(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userReq types.CreateUserRequest
		if err := c.ShouldBindJSON(&userReq); err != nil {
			c.Error(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.GetUserByAPIKey(c.Request.Context(), userReq.ApiKey)
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "API Key already in use"})
			return
		}

		user, err := db.CreateUser(c.Request.Context(), userReq)
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

func GetAuthToken(db *database.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var authTokenReq types.AuthTokenRequest
		if err := c.ShouldBindJSON(&authTokenReq); err != nil {
			c.Error(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		apiKeyUser, err := db.GetUserByAPIKey(c.Request.Context(), authTokenReq.ApiKey)
		if err != nil {
			c.Error(err)
			switch {
			case err == sql.ErrNoRows:
				c.JSON(http.StatusBadRequest, map[string]interface{}{"error": fmt.Sprintf("Invalid API Key: %s", authTokenReq.ApiKey)})
			default:
				c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Something went wrong"})
			}
			return
		}

		expiry := time.Now().Add(config.JWT_EXPIRY_DURATION_HOURS).Unix()

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"userUUID": apiKeyUser.UUID,
			"apiKey":   apiKeyUser.ApiKey,
			"exp":      expiry,
		})

		tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
		if err != nil {
			c.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.SetCookie("Authorization", tokenString, int((24*time.Hour.Seconds())*config.JWT_COOKIE_EXPIRY_DURATION_DAYS), "/", os.Getenv("DOMAIN"), false, true)

		c.Status(http.StatusOK)

	}
}
