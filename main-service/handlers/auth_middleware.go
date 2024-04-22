package handlers

import (
	"net/http"

	jwtkeys "main-service/jwt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func CheckAuth(c *gin.Context) {
	cookie, err := c.Cookie("jwt")
	if err != nil {
		c.Status(http.StatusUnauthorized)
		c.Writer.WriteString("No cookie")
		return
	}

	token, err := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
		return jwtkeys.JWTPublicKey, nil
	})
	if err != nil || !token.Valid {
		c.Status(http.StatusUnauthorized)
		c.Writer.WriteString("Invalid cookie")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.Status(http.StatusUnauthorized)
		c.Writer.WriteString("Invalid cookie")
		return
	}

	c.Set("userID", claims["id"])
}
