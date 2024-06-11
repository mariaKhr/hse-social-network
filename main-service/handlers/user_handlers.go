package handlers

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"main-service/db"
	jwtkeys "main-service/jwt"
	"main-service/schemas"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v5"
)

func Signup(c *gin.Context) {
	var creds schemas.UserCredentials
	if err := c.BindJSON(&creds); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if _, _, err := selectUserByLogin(creds.Login); err == nil {
		c.Status(http.StatusForbidden)
		c.Writer.WriteString("User already exist")
		return
	}

	if err := insertUserCreds(creds); err != nil {
		c.Status(http.StatusInternalServerError)
		c.Writer.WriteString(fmt.Sprintf("Error executing query: %v", err))
		return
	}

	userID, _, err := selectUserByLogin(creds.Login)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Writer.WriteString(fmt.Sprintf("Error executing query: %v", err))
		return
	}

	tokenString, err := signClaims(userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Writer.WriteString(fmt.Sprintf("Error signing claims: %v", err))
		return
	}
	c.SetCookie("jwt", tokenString, 24*60*60, "", "", false, true)
}

func Login(c *gin.Context) {
	var creds schemas.UserCredentials
	if err := c.BindJSON(&creds); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	userID, passwordHash, err := selectUserByLogin(creds.Login)
	if err == pgx.ErrNoRows || passwordHash != getPasswordHash(creds.Password) {
		c.Status(http.StatusForbidden)
		c.Writer.WriteString("Invalid login or/and password")
		return
	}

	tokenString, err := signClaims(userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Writer.WriteString(fmt.Sprintf("Error signing claims: %v", err))
		return
	}
	c.SetCookie("jwt", tokenString, 24*60*60, "", "", false, true)
}

func Profile(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user schemas.User
	if err := c.BindJSON(&user); err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	birthdate, _ := time.Parse("2006-01-02", user.Birthdate)

	_, err := db.Pool.Exec(context.Background(), `
	UPDATE users 
	SET (
		first_name,
		last_name,
		birthdate,
		email,
		phone_number
	) = ($1, $2, $3, $4, $5)
	WHERE id=$6
	`, user.FirstName, user.LastName, birthdate, user.Email, user.PhoneNumber, userID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		c.Writer.WriteString(fmt.Sprintf("Error executing query: %v", err))
		return
	}

	c.Status(http.StatusOK)
}

func getPasswordHash(password string) string {
	hash := md5.Sum([]byte(password))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func selectUserByLogin(login string) (uint64, string, error) {
	var userID uint64
	var passwordHash string
	err := db.Pool.QueryRow(
		context.Background(),
		"SELECT id, password_hash FROM users WHERE login=$1",
		login).Scan(&userID, &passwordHash)
	if err != nil {
		return 0, "", err
	}
	return userID, passwordHash, nil
}

func selectLoginByID(userID uint64) (string, error) {
	var login string
	err := db.Pool.QueryRow(
		context.Background(),
		"SELECT login FROM users WHERE id=$1",
		userID).Scan(&login)
	if err != nil {
		return "", err
	}
	return login, nil
}

func insertUserCreds(creds schemas.UserCredentials) error {
	_, err := db.Pool.Exec(
		context.Background(),
		"INSERT INTO users (login, password_hash) VALUES ($1, $2)",
		creds.Login,
		getPasswordHash(creds.Password),
	)
	return err
}

func signClaims(userID uint64) (string, error) {
	claims := &jwt.MapClaims{
		"id": userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, *claims)
	return token.SignedString(jwtkeys.JWTPrivateKey)
}
