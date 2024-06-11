package grpcclient

import (
	"log"
	pb "main-service/protos"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var PostService pb.PostServiceClient
var StatService pb.StatisticServiceClient

func InitGRPCClients() {
	postServiceConn, err := grpc.Dial(os.Getenv("POST_SERVER_ADDR"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	PostService = pb.NewPostServiceClient(postServiceConn)

	statServiceConn, err := grpc.Dial(os.Getenv("STATISTICS_SERVER_ADDR"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	StatService = pb.NewStatisticServiceClient(statServiceConn)
}
