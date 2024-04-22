package main

import (
	"fmt"
	"main-service/db"
	"main-service/handlers"
	jwtkeys "main-service/jwt"
	"os"

	grpcclient "main-service/grpc_client"

	"github.com/gin-gonic/gin"
)

func main() {
	jwtkeys.InitJWTKeys()

	db.InitDB()
	defer db.CloseDB()
	db.CreateTable()

	grpcclient.InitGRPCClient()

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

	router.Run(fmt.Sprintf(":%s", os.Getenv("PORT")))
}
