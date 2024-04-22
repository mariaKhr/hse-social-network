package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"post-service/db"
	pb "post-service/protos"
	"post-service/server"

	"google.golang.org/grpc"
)

func main() {
	db.InitDB()
	defer db.CloseDB()
	db.CreateTable()

	s := grpc.NewServer()
	pb.RegisterPostServiceServer(s, &server.PostServer{})

	address := fmt.Sprintf(":%v", os.Getenv("PORT"))
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintln(os.Stderr, "grpc server listening on", address)
	if err := s.Serve(lis); err != nil {
		log.Fatal(err)
	}
}
