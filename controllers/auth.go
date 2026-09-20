package controllers

import (
	"errors"
	"example/web-service-gin/auth"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid request",
		})
		return
	}

	claims, err := auth.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "invalid or expired refresh token",
		})
		return
	}

	tokens, nextState, err := auth.GenerateTokenPair(
		claims.UserID,
		claims.SessionID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to generate token",
		})
		return
	}

	err = auth.RotateSession(
		c.Request.Context(),
		claims,
		nextState,
	)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidSession) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "refresh token has expired or been used",
			})
			return
		}

		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "authentication service unavailable",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "token refreshed successfully",
		"token":   tokens,
	})
}
