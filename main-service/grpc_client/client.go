package grpcclient

import (
	"log"
	pb "main-service/protos"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var GRPCClient pb.PostServiceClient

func InitGRPCClient() {
	conn, err := grpc.Dial(os.Getenv("POST_SERVER_ADDR"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	GRPCClient = pb.NewPostServiceClient(conn)
}
