package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"

	"github.com/yusufatac/bitaksi-case-study/internal/domain"
	"github.com/yusufatac/bitaksi-case-study/internal/repository"
)

// Custom errors
var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
)

type AuthService interface {
	Register(ctx context.Context, creds domain.UserCredentials) (*domain.User, error)
	Login(ctx context.Context, username, password string) (string, error)
	ValidateToken(token string) (*Claims, error)
}

type Claims struct {
	UserID        string `json:"user_id"`
	Username      string `json:"username"`
	Authenticated bool   `json:"authenticated"`
	jwt.RegisteredClaims
}

type authService struct {
	userRepo repository.UserRepository
	jwtKey   []byte
}

func NewAuthService(userRepo repository.UserRepository, jwtSecret string) AuthService {
	return &authService{
		userRepo: userRepo,
		jwtKey:   []byte(jwtSecret),
	}
}

func (s *authService) Register(ctx context.Context, creds domain.UserCredentials) (*domain.User, error) {
	// Check if user exists
	existingUser, err := s.userRepo.GetUserByUsername(ctx, creds.Username)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Create new user
	user := domain.NewUser(creds)
	if err := user.HashPassword(); err != nil {
		return nil, err
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, username, password string) (string, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		return "", err
	}

	// Generate JWT token
	claims := &Claims{
		UserID:        user.ID,
		Username:      user.Username,
		Authenticated: true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *authService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid || !claims.Authenticated {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
