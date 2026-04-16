package auth

import (
	"booking/internal/domain"
	"booking/internal/utils"
	"booking/internal/utils/customErrors"
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type UserRepo interface {
	CreateUser(ctx context.Context, email, passwordHash, role string) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}

type TokenService interface {
	Generate(userID, role string) (string, error)
}

type Usecase struct {
	repo  UserRepo
	token TokenService
}

func New(repo UserRepo, token TokenService) *Usecase {
	return &Usecase{
		repo:  repo,
		token: token,
	}
}

func (u *Usecase) Register(ctx context.Context, email, password, role string) (*domain.User, error) {
	role = strings.ToLower(strings.TrimSpace(role))
	if role != utils.RoleAdmin && role != utils.RoleUser {
		return nil, customErrors.ErrInvalidRole
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := u.repo.CreateUser(ctx, strings.TrimSpace(email), string(hash), role)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *Usecase) Login(ctx context.Context, email, password string) (string, error) {
	user, err := u.repo.GetUserByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		if errors.Is(err, customErrors.ErrUserNotFound) {
			return "", customErrors.ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", customErrors.ErrInvalidCredentials
	}

	token, err := u.token.Generate(user.UserID, user.Role)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (u *Usecase) DummyLogin(role string) (string, error) {
	role = strings.ToLower(strings.TrimSpace(role))
	var userID string
	switch role {
	case utils.RoleUser:
		userID = "11111111-1111-1111-1111-111111111111"
	case utils.RoleAdmin:
		userID = "11111111-1111-1111-1111-111111111112"
	default:
		userID = "undef"
	}
	if userID == "undef" {
		return "", customErrors.ErrInvalidRole
	}
	token, err := u.token.Generate(userID, role)
	if err != nil {
		return "", err
	}
	return token, nil
}
