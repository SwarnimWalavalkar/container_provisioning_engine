package middlewares

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired(c *gin.Context) {
	cookie, err := c.Cookie("Authorization")
	if err != nil {
		c.Error(err)
		switch {
		case err == http.ErrNoCookie:
			c.AbortWithStatusJSON(http.StatusUnauthorized, map[string]interface{}{"error": "Authorization cookie missing"})
		default:
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	claims := jwt.MapClaims{}

	token, err := jwt.ParseWithClaims(cookie, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil || token == nil || !token.Valid {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Invalid auth token", "err": err.Error()})
		return
	}

	c.Set("userUUID", claims["userUUID"])

	c.Next()
}
