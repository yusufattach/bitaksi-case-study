package service

import (
	"context"
	"errors"
	"time"

	"github.com/yusufatac/bitaksi-case-study/internal/domain"
	"github.com/yusufatac/bitaksi-case-study/internal/repository"
)

type LocationService interface {
	UpdateDriverLocation(ctx context.Context, driverID string, lat, lon float64) error
	UpdateDriverLocations(ctx context.Context, locations []domain.DriverLocation) error
	FindNearbyDrivers(ctx context.Context, lat, lon, radius float64) ([]*domain.DriverLocation, error)
}

type locationService struct {
	repo repository.LocationRepository
}

func NewLocationService(repo repository.LocationRepository) LocationService {
	return &locationService{
		repo: repo,
	}
}

func (s *locationService) UpdateDriverLocation(ctx context.Context, driverID string, lat, lon float64) error {
	location := &domain.DriverLocation{
		DriverID:  driverID,
		Location:  domain.NewPoint(lat, lon),
		Status:    "active",
		Timestamp: time.Now(),
	}

	return s.repo.SaveLocation(ctx, location)
}

func (s *locationService) UpdateDriverLocations(ctx context.Context, locations []domain.DriverLocation) error {
	// Convert to pointer slice and set timestamps
	now := time.Now()
	locationPtrs := make([]*domain.DriverLocation, len(locations))
	for i := range locations {
		locations[i].Timestamp = now
		locationPtrs[i] = &locations[i]
	}

	return s.repo.SaveLocations(ctx, locationPtrs)
}

func (s *locationService) FindNearbyDrivers(ctx context.Context, lat, lon, radius float64) ([]*domain.DriverLocation, error) {
	// Validate input
	if lat < -90 || lat > 90 {
		return nil, ErrInvalidLatitude
	}
	if lon < -180 || lon > 180 {
		return nil, ErrInvalidLongitude
	}
	if radius <= 0 {
		return nil, ErrInvalidRadius
	}

	return s.repo.FindNearbyDrivers(ctx, lat, lon, radius)
}

// Custom errors
var (
	ErrInvalidLatitude  = errors.New("latitude must be between -90 and 90")
	ErrInvalidLongitude = errors.New("longitude must be between -180 and 180")
	ErrInvalidRadius    = errors.New("radius must be greater than 0")
)
