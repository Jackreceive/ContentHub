package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	issuer          = "jack"
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour

	accessTokenType  = "access"
	refreshTokenType = "refresh"
)

type Claims struct {
	UserID    int    `json:"user_id"`
	SessionID string `json:"session_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func NewSessionID() (string, error) {
	return randomID()
}

func GenerateTokenPair(
	userID int,
	sessionID string,
) (TokenPair, SessionState, error) {
	var pair TokenPair
	var state SessionState

	accessSecret, err := tokenSecret("JWT_ACCESS_SECRET")
	if err != nil {
		return pair, state, err
	}

	refreshSecret, err := tokenSecret("JWT_REFRESH_SECRET")
	if err != nil {
		return pair, state, err
	}

	accessJTI, err := randomID()
	if err != nil {
		return pair, state, err
	}

	refreshJTI, err := randomID()
	if err != nil {
		return pair, state, err
	}

	now := time.Now()

	accessClaims := Claims{
		UserID:    userID,
		SessionID: sessionID,
		TokenType: accessTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   strconv.Itoa(userID),
			ID:        accessJTI,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenTTL)),
		},
	}

	refreshClaims := Claims{
		UserID:    userID,
		SessionID: sessionID,
		TokenType: refreshTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    issuer,
			Subject:   strconv.Itoa(userID),
			ID:        refreshJTI,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshTokenTTL)),
		},
	}

	pair.AccessToken, err = signToken(accessClaims, accessSecret)
	if err != nil {
		return TokenPair{}, SessionState{}, err
	}

	pair.RefreshToken, err = signToken(refreshClaims, refreshSecret)
	if err != nil {
		return TokenPair{}, SessionState{}, err
	}

	state = SessionState{
		UserID:     userID,
		AccessJTI:  accessJTI,
		RefreshJTI: refreshJTI,
	}

	return pair, state, nil
}

func ParseAccessToken(tokenString string) (*Claims, error) {
	secret, err := tokenSecret("JWT_ACCESS_SECRET")
	if err != nil {
		return nil, err
	}

	return parseToken(tokenString, accessTokenType, secret)
}

func ParseRefreshToken(tokenString string) (*Claims, error) {
	secret, err := tokenSecret("JWT_REFRESH_SECRET")
	if err != nil {
		return nil, err
	}

	return parseToken(tokenString, refreshTokenType, secret)
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := extractBearerToken(
			c.GetHeader("Authorization"),
		)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "missing or invalid authorization header",
			})
			return
		}

		claims, err := ParseAccessToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid or expired access token",
			})
			return
		}

		err = ValidateAccessSession(c.Request.Context(), claims)
		if err != nil {
			if errors.Is(err, ErrInvalidSession) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"message": "session expired or revoked",
				})
				return
			}

			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"message": "authentication service unavailable",
			})
			return
		}

		c.Set("UserID", claims.UserID)
		c.Set("SessionID", claims.SessionID)
		c.Next()
	}
}

func parseToken(
	tokenString string,
	expectedType string,
	secret []byte,
) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(_ *jwt.Token) (any, error) {
			return secret, nil
		},
		jwt.WithValidMethods([]string{
			jwt.SigningMethodHS256.Alg(),
		}),
		jwt.WithIssuer(issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	if claims.UserID <= 0 ||
		claims.SessionID == "" ||
		claims.ID == "" ||
		claims.TokenType != expectedType {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func signToken(claims Claims, secret []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func tokenSecret(name string) ([]byte, error) {
	secret := os.Getenv(name)

	if len(secret) < 32 {
		return nil, errors.New(name + " must be at least 32 characters")
	}

	return []byte(secret), nil
}

func randomID() (string, error) {
	value := make([]byte, 16)

	if _, err := rand.Read(value); err != nil {
		return "", err
	}

	return hex.EncodeToString(value), nil
}

func extractBearerToken(authorization string) (string, error) {
	parts := strings.Fields(authorization)

	if len(parts) != 2 ||
		!strings.EqualFold(parts[0], "bearer") ||
		parts[1] == "" {
		return "", errors.New("invalid authorization header")
	}

	return parts[1], nil
}
