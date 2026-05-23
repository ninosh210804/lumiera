package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/ninosh210804/lumiera/apps/server/internal/domain"
)

type claims struct {
	jwt.RegisteredClaims
	UserID     string `json:"uid"`
	LocationID string `json:"loc"`
	Role       string `json:"role"`
}

// Issue creates a signed HS256 JWT for the given user.
func Issue(secret string, ttl time.Duration, userID, locationID uuid.UUID, role domain.RoleType) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(ttl)

	c := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
		UserID:     userID.String(),
		LocationID: locationID.String(),
		Role:       string(role),
	}

	tok, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("signing token: %w", err)
	}
	return tok, exp, nil
}

// Parse validates the token string and returns domain Claims.
func Parse(secret, tokenStr string) (*domain.Claims, error) {
	var c claims
	tok, err := jwt.ParseWithClaims(tokenStr, &c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, domain.ErrTokenInvalid
	}
	if !tok.Valid {
		return nil, domain.ErrTokenInvalid
	}

	userID, err := uuid.Parse(c.UserID)
	if err != nil {
		return nil, domain.ErrTokenInvalid
	}
	locationID, err := uuid.Parse(c.LocationID)
	if err != nil {
		return nil, domain.ErrTokenInvalid
	}

	exp := time.Time{}
	if c.ExpiresAt != nil {
		exp = c.ExpiresAt.Time
	}
	iat := time.Time{}
	if c.IssuedAt != nil {
		iat = c.IssuedAt.Time
	}

	return &domain.Claims{
		UserID:     userID,
		LocationID: locationID,
		Role:       domain.RoleType(c.Role),
		IssuedAt:   iat,
		ExpiresAt:  exp,
	}, nil
}
