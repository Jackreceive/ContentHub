package auth

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const issuer = "jack"
const tokenTTL = 24 * time.Hour

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func jwtSecret() ([]byte, error) {
	secret := "aadwadawdawdawjdawdjwadiawoidawodahwoduhawufdwadawdwadawd"
	if len(secret) < 32 {
		return nil, errors.New("JWT_SECRET must be at least 32 characters")
	}
	return []byte(secret), nil
}

func GenerateToken(userID int) (string, error) {
	secret, err := jwtSecret()
	if err != nil {
		return "", err
	}
	now := time.Now()
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   strconv.Itoa(userID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ParseToken(tokenString string) (*Claims, error) {
	secret, err := jwtSecret()
	if err != nil {
		return nil, err
	}
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(_ *jwt.Token) (any, error) {
			return secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}
	if claims.UserID <= 0 {
		return nil, errors.New("invalid user id")
	}
	return claims, nil
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.GetHeader("Authorization")
		parts := strings.Split(authorization, " ")
		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "bearer") ||
			strings.TrimSpace(parts[1]) == "" {
			c.AbortWithStatusJSON(401, gin.H{
				"message": "missing or invalid authorization header",
			})
			return
		}
		claims, err := ParseToken(authorization)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{
				"message": "invalid or expired authorization",
			})
			return
		}

		c.Set("UserID", claims.UserID)

		c.Next()

	}
}
