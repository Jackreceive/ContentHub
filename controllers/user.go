package controllers

import (
	"errors"
	"example/web-service-gin/db"
	"example/web-service-gin/models"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 创建此类别的目的是不让客户端提交 ID、时间戳等不应该由客户端控制的字段。
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Name     string `json:"name" binding:"required,min=2,max=50"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,min=2,max=50"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
		})
		return
	}
	var t models.User
	err := db.DB.
		Where("email=?", req.Email).First(&t).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to login",
			})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "invalid email or password",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(t.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "invalid email or password",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "login successfully",
		"user": gin.H{
			"id":    t.ID,
			"name":  t.Name,
			"email": t.Email,
		},
	})
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
		})
		return
	}
	email := strings.ToLower(req.Email)
	if err := db.DB.Select("id").Where("email=?", email).First(&models.User{}).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"message": "email already exists",
		})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to register",
		})
		return
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to register",
		})
		return
	}
	user := models.User{
		Name:     req.Name,
		Password: string(passwordHash),
		Email:    strings.ToLower(req.Email),
	}
	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to register",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "register successfully",
	})
}
