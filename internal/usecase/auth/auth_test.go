package auth

import (
	"booking/internal/domain"
	"booking/internal/utils/customErrors"
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type userRepoMock struct {
	createUserFn     func(ctx context.Context, email, passwordHash, role string) (*domain.User, error)
	getUserByEmailFn func(ctx context.Context, email string) (*domain.User, error)
}

func (m *userRepoMock) CreateUser(ctx context.Context, email, passwordHash, role string) (*domain.User, error) {
	if m.createUserFn == nil {
		return nil, errors.New("unexpected call CreateUser")
	}
	return m.createUserFn(ctx, email, passwordHash, role)
}

func (m *userRepoMock) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getUserByEmailFn == nil {
		return nil, errors.New("unexpected call GetUserByEmail")
	}
	return m.getUserByEmailFn(ctx, email)
}

type tokenMock struct {
	generateFn func(userID, role string) (string, error)
}

func (m tokenMock) Generate(userID, role string) (string, error) {
	if m.generateFn == nil {
		return "", errors.New("unexpected call Generate")
	}
	return m.generateFn(userID, role)
}

func TestRegister_RejectsInvalidRole(t *testing.T) {
	uc := New(&userRepoMock{}, tokenMock{})

	_, err := uc.Register(context.Background(), "a@b.c", "pass", "manager")
	if !errors.Is(err, customErrors.ErrInvalidRole) {
		t.Fatalf("expected ErrInvalidRole, got %v", err)
	}
}

func TestRegister_TrimsEmailAndHashesPassword(t *testing.T) {
	var gotEmail, gotHash, gotRole string
	repo := &userRepoMock{
		createUserFn: func(_ context.Context, email, passwordHash, role string) (*domain.User, error) {
			gotEmail = email
			gotHash = passwordHash
			gotRole = role
			return &domain.User{UserID: "u1", Email: email, Role: role}, nil
		},
	}

	uc := New(repo, tokenMock{})
	_, err := uc.Register(context.Background(), "  user@test.local  ", "pass123", "  USER  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotEmail != "user@test.local" {
		t.Fatalf("expected trimmed email, got %q", gotEmail)
	}
	if gotRole != "user" {
		t.Fatalf("expected normalized role user, got %q", gotRole)
	}
	if gotHash == "pass123" {
		t.Fatal("expected password to be hashed")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(gotHash), []byte("pass123")); err != nil {
		t.Fatalf("hash does not match original password: %v", err)
	}
}

func TestLogin_UserNotFoundToInvalidCredentials(t *testing.T) {
	repo := &userRepoMock{
		getUserByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, customErrors.ErrUserNotFound
		},
	}

	uc := New(repo, tokenMock{})
	_, err := uc.Login(context.Background(), "user@test.local", "pass")
	if !errors.Is(err, customErrors.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	repo := &userRepoMock{
		getUserByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return &domain.User{UserID: "u1", PasswordHash: string(hash), Role: "user"}, nil
		},
	}

	uc := New(repo, tokenMock{})
	_, err := uc.Login(context.Background(), "user@test.local", "wrong")
	if !errors.Is(err, customErrors.ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	repo := &userRepoMock{
		getUserByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return &domain.User{UserID: "u1", PasswordHash: string(hash), Role: "user"}, nil
		},
	}
	tokenSvc := tokenMock{generateFn: func(userID, role string) (string, error) {
		if userID != "u1" || role != "user" {
			return "", errors.New("unexpected token args")
		}
		return "jwt-token", nil
	}}

	uc := New(repo, tokenSvc)
	token, err := uc.Login(context.Background(), " user@test.local ", "correct")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "jwt-token" {
		t.Fatalf("expected token jwt-token, got %q", token)
	}
}

func TestDummyLogin_UnsupportedRole(t *testing.T) {
	uc := New(&userRepoMock{}, tokenMock{})

	_, err := uc.DummyLogin("manager")
	if !errors.Is(err, customErrors.ErrInvalidRole) {
		t.Fatalf("expected ErrInvalidRole, got %v", err)
	}
}

func TestDummyLogin_AdminRole(t *testing.T) {
	tokenSvc := tokenMock{generateFn: func(userID, role string) (string, error) {
		if userID != "11111111-1111-1111-1111-111111111112" {
			return "", errors.New("unexpected admin id")
		}
		if role != "admin" {
			return "", errors.New("unexpected role")
		}
		return "dummy-admin", nil
	}}

	uc := New(&userRepoMock{}, tokenSvc)
	token, err := uc.DummyLogin("ADMIN")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "dummy-admin" {
		t.Fatalf("expected token dummy-admin, got %q", token)
	}
}
