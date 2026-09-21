package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"example/web-service-gin/db"

	"github.com/redis/go-redis/v9"
)

const sessionKeyPrefix = "auth:session:"

var ErrInvalidSession = errors.New("invalid or expired session")

type SessionState struct {
	UserID     int    `json:"user_id"`
	AccessJTI  string `json:"access_jti"`
	RefreshJTI string `json:"refresh_jti"`
}

func StoreSession(
	ctx context.Context,
	sessionID string,
	state SessionState,
) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to encode session: %w", err)
	}

	err = db.Redis.Set(
		ctx,
		sessionID,
		data,
		refreshTokenTTL,
	).Err()
	if err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	return nil
}

func ValidateAccessSession(
	ctx context.Context,
	claims *Claims,
) error {
	data, err := db.Redis.Get(
		ctx,
		claims.SessionID,
	).Bytes()

	if errors.Is(err, redis.Nil) {
		return ErrInvalidSession
	}
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}

	var state SessionState
	if err := json.Unmarshal(data, &state); err != nil {
		return ErrInvalidSession
	}

	if state.UserID != claims.UserID ||
		state.AccessJTI != claims.ID {
		return ErrInvalidSession
	}

	return nil
}

func RotateSession(
	ctx context.Context,
	claims *Claims,
	next SessionState,
) error {
	key := claims.SessionID

	nextData, err := json.Marshal(next)
	if err != nil {
		return fmt.Errorf("failed to encode session: %w", err)
	}

	err = db.Redis.Watch(
		ctx,
		func(tx *redis.Tx) error {
			data, err := tx.Get(ctx, key).Bytes()
			if err != nil {
				return err
			}

			var current SessionState
			if err := json.Unmarshal(data, &current); err != nil {
				return ErrInvalidSession
			}

			if current.UserID != claims.UserID ||
				current.RefreshJTI != claims.ID {
				return ErrInvalidSession
			}

			_, err = tx.TxPipelined(
				ctx,
				func(pipe redis.Pipeliner) error {
					pipe.Set(
						ctx,
						key,
						nextData,
						refreshTokenTTL,
					)
					return nil
				},
			)

			return err
		},
		key,
	)

	if errors.Is(err, redis.Nil) ||
		errors.Is(err, redis.TxFailedErr) ||
		errors.Is(err, ErrInvalidSession) {
		return ErrInvalidSession
	}

	if err != nil {
		return fmt.Errorf("failed to rotate session: %w", err)
	}

	return nil
}

func RevokeSession(
	ctx context.Context,
	sessionID string,
) error {
	if err := db.Redis.Del(
		ctx,
		sessionID,
	).Err(); err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	return nil
}

func sessionKey(sessionID string) string {
	return sessionKeyPrefix + sessionID
}
