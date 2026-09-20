package controllers

import (
	"errors"
	"example/web-service-gin/db"
	"example/web-service-gin/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ArticleRequest struct {
	Title   string `json:"title" binding:"required,max=100"`
	Content string `json:"content" binding:"required"`
}

func CreateArticle(c *gin.Context) {
	userID := c.GetInt("UserID")

	var req ArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
		})
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" || strings.TrimSpace(req.Content) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "title and content cannot be empty",
		})
		return
	}

	article := models.Article{
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
	}

	if err := db.DB.Create(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create article",
		})
		return
	}

	if err := db.DB.Preload("User").First(&article, article.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "article created, but failed to load article",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "article created successfully",
		"article": article,
	})
}

func ArticleAll(c *gin.Context) {
	var articles []models.Article

	if err := db.DB.
		Preload("User").
		Order("created_at DESC").
		Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get articles",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"articles": articles,
	})
}

func GetArticleByID(c *gin.Context) {
	id, ok := getArticleID(c)
	if !ok {
		return
	}

	var article models.Article
	err := db.DB.
		Preload("User").
		First(&article, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "article not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get article",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"article": article,
	})
}

func UpdateArticleByID(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	id, ok := getArticleID(c)
	if !ok {
		return
	}

	var req ArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
		})
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" || strings.TrimSpace(req.Content) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "title and content cannot be empty",
		})
		return
	}

	var article models.Article
	err := db.DB.
		Where("id = ? AND user_id = ?", id, userID).
		First(&article).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "article not found",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to get article",
		})
		return
	}

	article.Title = req.Title
	article.Content = req.Content

	if err := db.DB.Save(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to update article",
		})
		return
	}

	if err := db.DB.Preload("User").First(&article, article.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "article updated, but failed to load article",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "article updated successfully",
		"article": article,
	})
}

func DeleteArticleByID(c *gin.Context) {
	userID, ok := getCurrentUserID(c)
	if !ok {
		return
	}

	id, ok := getArticleID(c)
	if !ok {
		return
	}

	result := db.DB.
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.Article{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to delete article",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "article not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "article deleted successfully",
	})
}

func getArticleID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid article id",
		})
		return 0, false
	}

	return id, true
}

func getCurrentUserID(c *gin.Context) (int, bool) {
	value, exists := c.Get("UserID")
	userID, ok := value.(int)

	if !exists || !ok || userID <= 0 {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return 0, false
	}

	return userID, true
}
