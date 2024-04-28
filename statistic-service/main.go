package main

import (
	"statistic-service/db"
	"statistic-service/handlers"
	kafka "statistic-service/kafka_consumer"

	"github.com/gin-gonic/gin"
)

func main() {
	db.InitDB()
	go kafka.RunKafkaConsumer()

	router := gin.Default()

	router.GET("/ping", handlers.Ping)
}
