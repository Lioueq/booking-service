package jwtservice

import (
	"booking/internal/utils/customErrors"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secretKey []byte
	ttl       time.Duration
}

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func New(secret string, ttl time.Duration) *JWTService {
	if secret == "" {
		secret = "secret"
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	return &JWTService{
		secretKey: []byte(secret),
		ttl:       ttl,
	}
}

func (s *JWTService) Generate(userID, role string) (string, error) {
	now := time.Now()

	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

func (s *JWTService) Parse(tokenStr string) (string, string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, customErrors.ErrUnauthorized
		}
		return s.secretKey, nil
	})
	if err != nil {
		return "", "", customErrors.ErrUnauthorized
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return "", "", customErrors.ErrUnauthorized
	}

	if claims.Subject == "" || claims.Role == "" {
		return "", "", customErrors.ErrUnauthorized
	}

	return claims.Subject, claims.Role, nil
}

func (s *JWTService) ParseToken(tokenStr string) (string, string, error) {
	userID, role, err := s.Parse(tokenStr)
	if err != nil {
		if errors.Is(err, customErrors.ErrUnauthorized) {
			return "", "", err
		}
		return "", "", customErrors.ErrUnauthorized
	}
	return userID, role, nil
}
