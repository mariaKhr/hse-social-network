package main

import (
	"fmt"
	"main-service/db"
	"main-service/handlers"
	jwtkeys "main-service/jwt"
	kafka "main-service/kafka_producer"
	"os"

	grpcclient "main-service/grpc_client"

	"github.com/gin-gonic/gin"
)

func main() {
	jwtkeys.InitJWTKeys()

	db.InitDB()
	defer db.CloseDB()
	db.CreateTable()

	grpcclient.InitGRPCClients()
	kafka.InitKafkaProducer()
	defer kafka.CloseKafkaProducer()

	router := gin.Default()

	user := router.Group("/user")
	{
		user.POST("/signup", handlers.Signup)
		user.POST("/login", handlers.Login)
		user.PUT("/profile", handlers.CheckAuth, handlers.Profile)
	}

	post := router.Group("/post")
	{
		post.POST("", handlers.CheckAuth, handlers.CreatePost)
		post.PUT("/:id", handlers.CheckAuth, handlers.UpdatePost)
		post.GET("/:id", handlers.CheckAuth, handlers.GetPost)
		post.DELETE("/:id", handlers.CheckAuth, handlers.DeletePost)
		post.GET("/page", handlers.CheckAuth, handlers.GetPage)
	}

	router.POST("/like/:id", handlers.CheckAuth, handlers.LikePost)
	router.POST("/view/:id", handlers.CheckAuth, handlers.ViewPost)

	stat := router.Group("/stat")
	{
		stat.GET("/:id", handlers.CheckAuth, handlers.GetStatisticsByPost)
		stat.GET("/top5posts", handlers.CheckAuth, handlers.GetTop5Posts)
		stat.GET("/top3users", handlers.CheckAuth, handlers.GetTop3Users)
	}

	router.Run(fmt.Sprintf(":%s", os.Getenv("PORT")))
}
