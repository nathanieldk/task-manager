package usecase

import (
	"context"
	"time"

	"github.com/nathanieldk/task-manager/internal/dto"
	"github.com/nathanieldk/task-manager/internal/entity"
	"github.com/nathanieldk/task-manager/internal/pkg/apperror"
	"github.com/nathanieldk/task-manager/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error)
}

type authUsecase struct {
	userRepo  repository.UserRepository
	jwtSecret string
	jwtExpiry time.Duration
}

func NewAuthUsecase(
	userRepo repository.UserRepository,
	jwtSecret string,
	jwtExpiry time.Duration,
) AuthUsecase {
	return &authUsecase{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

func (uc *authUsecase) Register(ctx context.Context, req dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Check email uniqueness
	existing, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.ErrInternal(err)
	}
	if existing != nil {
		return nil, apperror.ErrConflict("Email already registered")
	}

	// Check username uniqueness
	existing, err = uc.userRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, apperror.ErrInternal(err)
	}
	if existing != nil {
		return nil, apperror.ErrConflict("Username already taken")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.ErrInternal(err)
	}

	now := time.Now().UTC()
	user := &entity.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		TeamID:       req.TeamID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return nil, apperror.ErrInternal(err)
	}

	token, err := uc.generateToken(user)
	if err != nil {
		return nil, apperror.ErrInternal(err)
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserInfo{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			TeamID:   user.TeamID,
		},
	}, nil
}

func (uc *authUsecase) Login(ctx context.Context, req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := uc.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.ErrInternal(err)
	}
	if user == nil {
		return nil, apperror.ErrUnauthorized("Invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, apperror.ErrUnauthorized("Invalid email or password")
	}

	token, err := uc.generateToken(user)
	if err != nil {
		return nil, apperror.ErrInternal(err)
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserInfo{
			ID:       user.ID,
			Email:    user.Email,
			Username: user.Username,
			TeamID:   user.TeamID,
		},
	}, nil
}

func (uc *authUsecase) generateToken(user *entity.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"team_id": user.TeamID,
		"exp":     time.Now().Add(uc.jwtExpiry).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(uc.jwtSecret))
}
