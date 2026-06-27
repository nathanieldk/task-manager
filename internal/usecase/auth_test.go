package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/nathanieldk/task-manager/internal/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAuthUsecase() (AuthUsecase, *mockUserRepo) {
	userRepo := newMockUserRepo()
	uc := NewAuthUsecase(userRepo, "test-jwt-secret", 24*time.Hour)
	return uc, userRepo
}

func TestRegister_Success(t *testing.T) {
	uc, userRepo := newTestAuthUsecase()
	ctx := context.Background()

	req := dto.RegisterRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
		TeamID:   1,
	}

	result, err := uc.Register(ctx, req)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Token)
	assert.Equal(t, "test@example.com", result.User.Email)
	assert.Equal(t, "testuser", result.User.Username)
	assert.Equal(t, 1, result.User.TeamID)

	// Verify user was stored
	assert.Len(t, userRepo.users, 1)
}

func TestRegister_DuplicateEmail(t *testing.T) {
	uc, _ := newTestAuthUsecase()
	ctx := context.Background()

	req := dto.RegisterRequest{
		Email:    "dupe@example.com",
		Username: "user1",
		Password: "password123",
		TeamID:   1,
	}

	_, err := uc.Register(ctx, req)
	require.NoError(t, err)

	// Register again with the same email
	req.Username = "user2"
	_, err = uc.Register(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Email already registered")
}

func TestRegister_DuplicateUsername(t *testing.T) {
	uc, _ := newTestAuthUsecase()
	ctx := context.Background()

	req := dto.RegisterRequest{
		Email:    "user1@example.com",
		Username: "sameuser",
		Password: "password123",
		TeamID:   1,
	}

	_, err := uc.Register(ctx, req)
	require.NoError(t, err)

	// Register again with the same username
	req.Email = "user2@example.com"
	_, err = uc.Register(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Username already taken")
}

func TestLogin_Success(t *testing.T) {
	uc, _ := newTestAuthUsecase()
	ctx := context.Background()

	// First register
	regReq := dto.RegisterRequest{
		Email:    "login@example.com",
		Username: "loginuser",
		Password: "password123",
		TeamID:   1,
	}
	_, err := uc.Register(ctx, regReq)
	require.NoError(t, err)

	// Then login
	loginReq := dto.LoginRequest{
		Email:    "login@example.com",
		Password: "password123",
	}
	result, err := uc.Login(ctx, loginReq)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Token)
	assert.Equal(t, "login@example.com", result.User.Email)
}

func TestLogin_WrongPassword(t *testing.T) {
	uc, _ := newTestAuthUsecase()
	ctx := context.Background()

	// Register
	regReq := dto.RegisterRequest{
		Email:    "wrong@example.com",
		Username: "wronguser",
		Password: "correctpassword",
		TeamID:   1,
	}
	_, err := uc.Register(ctx, regReq)
	require.NoError(t, err)

	// Login with wrong password
	loginReq := dto.LoginRequest{
		Email:    "wrong@example.com",
		Password: "wrongpassword",
	}
	_, err = uc.Login(ctx, loginReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid email or password")
}

func TestLogin_NonexistentUser(t *testing.T) {
	uc, _ := newTestAuthUsecase()
	ctx := context.Background()

	loginReq := dto.LoginRequest{
		Email:    "noone@example.com",
		Password: "password123",
	}
	_, err := uc.Login(ctx, loginReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Invalid email or password")
}
