package main

import (
	"crypto/rsa"
	"os"

	"github.com/golang-jwt/jwt"
)

func ParseJWTPrivateKey(jwtPrivateFile string) (*rsa.PrivateKey, error) {
	private, err := os.ReadFile(jwtPrivateFile)
	if err != nil {
		return nil, err
	}
	jwtPrivate, err := jwt.ParseRSAPrivateKeyFromPEM(private)
	if err != nil {
		return nil, err
	}
	return jwtPrivate, nil
}

func ParseJWTPublicKey(jwtPublicFile string) (*rsa.PublicKey, error) {
	public, err := os.ReadFile(jwtPublicFile)
	if err != nil {
		return nil, err
	}
	jwtPublic, err := jwt.ParseRSAPublicKeyFromPEM(public)
	if err != nil {
		return nil, err
	}
	return jwtPublic, nil
}
