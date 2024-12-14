package domain

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a system user (driver or rider)
type User struct {
	ID          string    `json:"id" bson:"_id,omitempty"`
	Username    string    `json:"username" bson:"username"`
	Password    string    `json:"-" bson:"password"`
	Email       string    `json:"email" bson:"email"`
	Status      string    `json:"status" bson:"status"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	LastLoginAt time.Time `json:"last_login_at" bson:"last_login_at"`
}

// UserCredentials represents login/register credentials
type UserCredentials struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6"`
	Email    string `json:"email" validate:"required,email"`
}

// HashPassword creates a bcrypt hash of the password
func (u *User) HashPassword() error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// ComparePassword compares the hash with plain text password
func (u *User) ComparePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// NewUser creates a new user with default values
func NewUser(creds UserCredentials) *User {
	now := time.Now()
	return &User{
		Username:    creds.Username,
		Password:    creds.Password,
		Email:       creds.Email,
		Status:      "active",
		CreatedAt:   now,
		LastLoginAt: now,
	}
}
