package main

import (
	"context"
	"crypto/md5"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx"
)

type UserCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func PasswordHash(password string) string {
	hash := md5.Sum([]byte(password))
	return base64.StdEncoding.EncodeToString(hash[:])
}

type AuthHandlers struct {
	jwtPrivate *rsa.PrivateKey
	jwtPublic  *rsa.PublicKey
	db         *UsersDatabase
}

func NewAuthHandlers(jwtPrivate *rsa.PrivateKey, jwtPublic *rsa.PublicKey, db *UsersDatabase) *AuthHandlers {
	return &AuthHandlers{
		jwtPrivate: jwtPrivate,
		jwtPublic:  jwtPublic,
		db:         db,
	}
}

func (h *AuthHandlers) Signup(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "signup can be done only with POST HTTP method")
		return
	}

	creds, err := NewUserCredentials(req)

	var httpErr *HttpError
	if errors.As(err, &httpErr) {
		w.WriteHeader(httpErr.Status)
		fmt.Fprint(w, httpErr.Msg)
		return
	}

	var userId int
	err = h.db.pool.QueryRow(
		context.Background(),
		"SELECT id FROM users WHERE login=$1",
		creds.Login).Scan(&userId)

	if err == nil {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "User already exists")
		return
	}

	_, err = h.db.pool.Exec(
		context.Background(),
		"INSERT INTO users (login, password_hash) VALUES ($1, $2)",
		creds.Login,
		PasswordHash(creds.Password),
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error executing query: %v", err)
		return
	}

	h.db.pool.QueryRow(
		context.Background(),
		"SELECT id FROM users WHERE login=$1",
		creds.Login).Scan(&userId)

	h.SetCookie(w, &jwt.MapClaims{
		"id": userId,
	})
}

func (h *AuthHandlers) Login(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "login can be done only with POST HTTP method")
		return
	}

	creds, err := NewUserCredentials(req)

	var httpErr *HttpError
	if errors.As(err, &httpErr) {
		w.WriteHeader(httpErr.Status)
		fmt.Fprint(w, httpErr.Msg)
		return
	}

	var userId int
	var passwordHash string
	err = h.db.pool.QueryRow(
		context.Background(),
		"SELECT id, password_hash FROM users WHERE login=$1",
		creds.Login).Scan(&userId, &passwordHash)

	if err == pgx.ErrNoRows || passwordHash != PasswordHash(creds.Password) {
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "Invalid credentials")
		return
	}

	h.SetCookie(w, &jwt.MapClaims{
		"id": userId,
	})
}

func NewUserCredentials(req *http.Request) (*UserCredentials, error) {
	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, &HttpError{
			Msg:    fmt.Sprintf("Error reading body: %v", err),
			Status: http.StatusInternalServerError,
		}
	}
	creds := UserCredentials{}
	err = json.Unmarshal(body, &creds)
	if err != nil {
		return nil, &HttpError{
			Msg:    fmt.Sprintf("Error unmarshalling body: %v", err),
			Status: http.StatusBadRequest,
		}
	}
	return &creds, nil
}

func (h *AuthHandlers) SetCookie(w http.ResponseWriter, claims *jwt.MapClaims) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, *claims)
	tokenString, err := token.SignedString(h.jwtPrivate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error signing token: %v", err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "jwt",
		Value: tokenString,
	})
}
