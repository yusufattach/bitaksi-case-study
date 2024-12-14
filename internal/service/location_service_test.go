package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yusufatac/bitaksi-case-study/internal/domain"
)

// MockLocationRepository is a mock implementation of the LocationRepository interface
type MockLocationRepository struct {
	mock.Mock
}

func (m *MockLocationRepository) SaveLocation(ctx context.Context, location *domain.DriverLocation) error {
	args := m.Called(ctx, location)
	return args.Error(0)
}

func (m *MockLocationRepository) SaveLocations(ctx context.Context, locations []*domain.DriverLocation) error {
	args := m.Called(ctx, locations)
	return args.Error(0)
}

func (m *MockLocationRepository) FindNearbyDrivers(ctx context.Context, lat, lon, radius float64) ([]*domain.DriverLocation, error) {
	args := m.Called(ctx, lat, lon, radius)
	return args.Get(0).([]*domain.DriverLocation), args.Error(1)
}

func TestUpdateDriverLocation(t *testing.T) {
	mockRepo := new(MockLocationRepository)
	service := NewLocationService(mockRepo)

	driverID := "driver1"
	lat, lon := 40.7128, -74.0060
	location := &domain.DriverLocation{
		DriverID:  driverID,
		Location:  domain.NewPoint(lat, lon),
		Status:    "active",
		Timestamp: time.Now(),
	}

	mockRepo.On("SaveLocation", mock.Anything, location).Return(nil)

	err := service.UpdateDriverLocation(context.Background(), driverID, lat, lon)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUpdateDriverLocations(t *testing.T) {
	mockRepo := new(MockLocationRepository)
	service := NewLocationService(mockRepo)

	locations := []domain.DriverLocation{
		{
			DriverID:  "driver1",
			Location:  domain.NewPoint(40.7128, -74.0060),
			Status:    "active",
			Timestamp: time.Now(),
		},
		{
			DriverID:  "driver2",
			Location:  domain.NewPoint(34.0522, -118.2437),
			Status:    "active",
			Timestamp: time.Now(),
		},
	}

	locationPtrs := []*domain.DriverLocation{
		&locations[0],
		&locations[1],
	}

	mockRepo.On("SaveLocations", mock.Anything, locationPtrs).Return(nil)

	err := service.UpdateDriverLocations(context.Background(), locations)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestFindNearbyDrivers(t *testing.T) {
	mockRepo := new(MockLocationRepository)
	service := NewLocationService(mockRepo)

	lat, lon, radius := 40.7128, -74.0060, 10.0
	drivers := []*domain.DriverLocation{
		{
			DriverID:  "driver1",
			Location:  domain.NewPoint(40.7128, -74.0060),
			Status:    "active",
			Timestamp: time.Now(),
		},
	}

	mockRepo.On("FindNearbyDrivers", mock.Anything, lat, lon, radius).Return(drivers, nil)

	result, err := service.FindNearbyDrivers(context.Background(), lat, lon, radius)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, drivers, result)
	mockRepo.AssertExpectations(t)
}
