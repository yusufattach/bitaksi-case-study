package service

import (
	"context"
	"golang.org/x/crypto/bcrypt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yusufatac/bitaksi-case-study/internal/domain"
)

// MockUserRepository is a mock implementation of the UserRepository interface
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestRegister(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-secret")

	creds := domain.UserCredentials{
		Username: "testuser",
		Password: "password",
	}

	mockRepo.On("GetUserByUsername", mock.Anything, creds.Username).Return(nil, nil)
	mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	user, err := service.Register(context.Background(), creds)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, creds.Username, user.Username)
	mockRepo.AssertExpectations(t)
}

func TestLogin(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-secret")

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := &domain.User{
		ID:       "1",
		Username: "testuser",
		Password: string(hashedPassword),
	}

	mockRepo.On("GetUserByUsername", mock.Anything, user.Username).Return(user, nil)
	mockRepo.On("UpdateLastLogin", mock.Anything, user.ID).Return(nil)

	token, err := service.Login(context.Background(), user.Username, "password")

	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	mockRepo.AssertExpectations(t)
}

func TestValidateToken(t *testing.T) {
	mockRepo := new(MockUserRepository)
	service := NewAuthService(mockRepo, "test-secret")

	claims := &Claims{
		UserID:        "1",
		Username:      "testuser",
		Authenticated: true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("test-secret"))

	validatedClaims, err := service.ValidateToken(tokenString)

	assert.NoError(t, err)
	assert.NotNil(t, validatedClaims)
	assert.Equal(t, claims.UserID, validatedClaims.UserID)
	assert.Equal(t, claims.Username, validatedClaims.Username)
	mockRepo.AssertExpectations(t)
}
