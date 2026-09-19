package main

import (
	"example/web-service-gin/controllers"
	"example/web-service-gin/db"
	"example/web-service-gin/models"

	"github.com/gin-gonic/gin"
)

func main() {
	err := db.InitMySQL()
	if err != nil {
		panic(err)
	}
	db.DB.AutoMigrate(&models.User{})
	r := gin.Default()
	v1 := r.Group("/v1")
	{
		v1.POST("user/login", controllers.Login)
		v1.POST("/user/register", controllers.Register)
	}
	r.Run(":9090")
}
