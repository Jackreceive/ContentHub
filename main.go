package main

import (
	"context"
	"errors"
	"example/web-service-gin/auth"
	"example/web-service-gin/controllers"
	"example/web-service-gin/db"
	"example/web-service-gin/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		panic(errors.New("need .env file"))
	}
	err := db.InitMySQL()
	if err != nil {
		panic(err)
	}
	if err := db.DB.AutoMigrate(&models.User{}, &models.Article{}); err != nil {
		panic(err)
	}

	ctx := context.Background()

	if err := db.InitRedis(ctx); err != nil {
		panic(err)
	}
	defer db.CloseRedis()

	r := gin.Default()
	v1 := r.Group("/v1")
	{
		v1.POST("user/login", controllers.Login)
		v1.POST("/user/register", controllers.Register)
		v1.POST("/token/refresh", controllers.RefreshToken)

		v1.Use(auth.Middleware())
		v1.GET("/user/logout", controllers.Logout)
		v1.POST("/article", controllers.CreateArticle)
		v1.GET("/articles", controllers.ArticleAll)
		v1.GET("/article/:id", controllers.GetArticleByID)
		v1.PUT("/article/:id", controllers.UpdateArticleByID)
		v1.DELETE("/article/:id", controllers.DeleteArticleByID)
	}
	if err := r.Run(":9090"); err != nil {
		panic(err)
	}
}
