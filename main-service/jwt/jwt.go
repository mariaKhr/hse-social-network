package jwtkeys

import (
	"crypto/rsa"
	"log"
	"os"

	"github.com/golang-jwt/jwt"
)

var JWTPrivateKey *rsa.PrivateKey
var JWTPublicKey *rsa.PublicKey

func InitJWTKeys() {
	initJWTPrivateKey()
	initJWTPublicKey()
}

func initJWTPrivateKey() {
	jwtPrivateFile := os.Getenv("JWT_PRIVATE_KEY_FILE")
	private, err := os.ReadFile(jwtPrivateFile)
	if err != nil {
		log.Fatal("Error getting jwt private key:", err)
	}

	JWTPrivateKey, err = jwt.ParseRSAPrivateKeyFromPEM(private)
	if err != nil {
		log.Fatal("Error getting jwt private key:", err)
	}
}

func initJWTPublicKey() {
	jwtPublicFile := os.Getenv("JWT_PUBLIC_KEY_FILE")
	public, err := os.ReadFile(jwtPublicFile)
	if err != nil {
		log.Fatal("Error getting jwt public key:", err)
	}

	JWTPublicKey, err = jwt.ParseRSAPublicKeyFromPEM(public)
	if err != nil {
		log.Fatal("Error getting jwt public key:", err)
	}
}
