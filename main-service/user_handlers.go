package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

type UserInfo struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Birthdate   string `json:"birthdate"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phoneNumber"`
}

type UserHandlers struct {
	jwtPublic *rsa.PublicKey
	db        *UsersDatabase
}

func NewUserHandlers(jwtPublic *rsa.PublicKey, db *UsersDatabase) *UserHandlers {
	return &UserHandlers{
		jwtPublic: jwtPublic,
		db:        db,
	}
}

func (h *UserHandlers) Profile(w http.ResponseWriter, req *http.Request) {
	claims, err := h.GetClaimsFromCookie(req)

	var httpErr *HttpError
	if errors.As(err, &httpErr) {
		w.WriteHeader(httpErr.Status)
		fmt.Fprint(w, httpErr.Msg)
		return
	}

	userId := int((*claims)["id"].(float64))

	info, err := NewUserInfo(req)
	if errors.As(err, &httpErr) {
		w.WriteHeader(httpErr.Status)
		fmt.Fprint(w, httpErr.Msg)
		return
	}

	birthdate, _ := time.Parse("2006-01-02", info.Birthdate)

	h.db.pool.Exec(context.Background(), `
	UPDATE users SET (first_name, last_name, birthdate, email, phone_number) = ($1, $2, $3, $4, $5)
		WHERE id=$6
	`, info.FirstName, info.LastName, birthdate, info.Email, info.PhoneNumber, userId)
}

func (h *UserHandlers) GetClaimsFromCookie(req *http.Request) (*jwt.MapClaims, error) {
	cookie, err := req.Cookie("jwt")
	if err != nil {
		return nil, &HttpError{
			Msg:    "Unauthorized",
			Status: http.StatusUnauthorized,
		}
	}

	tokenString := cookie.Value
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return h.jwtPublic, nil
	})
	if err != nil || !token.Valid {
		return nil, &HttpError{
			Msg:    "Unauthorized",
			Status: http.StatusBadRequest,
		}
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, &HttpError{
			Msg:    "Error getting claims",
			Status: http.StatusInternalServerError,
		}
	}
	return &claims, nil
}

func NewUserInfo(req *http.Request) (*UserInfo, error) {
	defer req.Body.Close()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, &HttpError{
			Msg:    fmt.Sprintf("Error reading body: %v", err),
			Status: http.StatusInternalServerError,
		}
	}
	info := UserInfo{}
	err = json.Unmarshal(body, &info)
	if err != nil {
		return nil, &HttpError{
			Msg:    fmt.Sprintf("Error unmarshalling body: %v", err),
			Status: http.StatusBadRequest,
		}
	}
	return &info, nil
}
