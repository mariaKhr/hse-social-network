package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"statistic-service/db"
	grpcserver "statistic-service/grpc_server"
	"statistic-service/handlers"
	kafka "statistic-service/kafka_consumer"
	pb "statistic-service/protos"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	db.InitDB()

	go kafka.RunKafkaConsumer("likes")
	go kafka.RunKafkaConsumer("views")

	go func() {
		s := grpc.NewServer()
		pb.RegisterStatisticServiceServer(s, &grpcserver.StatServer{})

		address := fmt.Sprintf(":%s", os.Getenv("GRPC_PORT"))
		lis, err := net.Listen("tcp", address)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Fprintln(os.Stderr, "grpc server listening on", address)
		if err := s.Serve(lis); err != nil {
			log.Fatal(err)
		}
	}()

	router := gin.Default()

	router.GET("/ping", handlers.Ping)

	router.Run(fmt.Sprintf(":%s", os.Getenv("HTTP_PORT")))
}
