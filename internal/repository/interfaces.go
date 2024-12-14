package repository

import (
	"context"

	"github.com/yusufatac/bitaksi-case-study/internal/domain"
)

// LocationRepository defines the interface for driver location operations
type LocationRepository interface {
	// SaveLocation saves or updates a driver's location
	SaveLocation(ctx context.Context, location *domain.DriverLocation) error

	// SaveLocations saves multiple driver locations in batch
	SaveLocations(ctx context.Context, locations []*domain.DriverLocation) error

	// FindNearbyDrivers finds drivers within a specified radius
	FindNearbyDrivers(ctx context.Context, lat, lon, radius float64) ([]*domain.DriverLocation, error)
}

// UserRepository defines the interface for user operations
type UserRepository interface {
	// CreateUser creates a new user
	CreateUser(ctx context.Context, user *domain.User) error

	// GetUserByUsername retrieves a user by username
	GetUserByUsername(ctx context.Context, username string) (*domain.User, error)

	// UpdateLastLogin updates the last login timestamp
	UpdateLastLogin(ctx context.Context, userID string) error
}
