package main

import (
	"fmt"
	"os"
	"statistic-service/db"
	"statistic-service/handlers"
	kafka "statistic-service/kafka_consumer"

	"github.com/gin-gonic/gin"
)

func main() {
	db.InitDB()
	go kafka.RunKafkaConsumer("likes")
	go kafka.RunKafkaConsumer("views")

	router := gin.Default()

	router.GET("/ping", handlers.Ping)

	router.Run(fmt.Sprintf(":%s", os.Getenv("PORT")))
}
