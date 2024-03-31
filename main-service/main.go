package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/kouhin/envflag"
)

func main() {
	privateFile := flag.String("jwt-private-key-file", "", "path to JWT private key `file`")
	publicFile := flag.String("jwt-public-key-file", "", "path to JWT public key `file`")
	port := flag.Int("port", 8091, "http server port")
	databaseUrl := flag.String("database-url", "", "database URL in format `postgresql://username:password@localhost:5432/database_name`")
	envflag.Parse()

	if port == nil {
		fmt.Fprintln(os.Stderr, "Port is required")
		os.Exit(1)
	}
	if privateFile == nil || *privateFile == "" {
		fmt.Fprintln(os.Stderr, "Please provide a path to JWT private key file")
		os.Exit(1)
	}
	if publicFile == nil || *publicFile == "" {
		fmt.Fprintln(os.Stderr, "Please provide a path to JWT public key file")
		os.Exit(1)
	}
	if databaseUrl == nil || *databaseUrl == "" {
		fmt.Fprintln(os.Stderr, "Please provide a database URL")
		os.Exit(1)
	}

	absoluteprivateFile, err := filepath.Abs(*privateFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	absolutePublicFile, err := filepath.Abs(*publicFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	jwtPrivate, err := ParseJWTPrivateKey(absoluteprivateFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	jwtPublic, err := ParseJWTPublicKey(absolutePublicFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	db := NewUsersDatabase(*databaseUrl)
	defer db.Close()
	err = db.CreateTable()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	authHandlers := NewAuthHandlers(jwtPrivate, jwtPublic, db)
	http.HandleFunc("/signup", authHandlers.Signup)
	http.HandleFunc("/login", authHandlers.Login)

	userHandlers := NewUserHandlers(jwtPublic, db)
	http.HandleFunc("/profile", userHandlers.Profile)

	fmt.Println("Starting server on port", *port, "with jwt private key file", absoluteprivateFile, "and jwt public key file", absolutePublicFile)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
