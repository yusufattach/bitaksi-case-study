package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yusufatac/bitaksi-case-study/internal/domain"
)

// MockLocationService is a mock implementation of the LocationService interface
type MockLocationService struct {
	mock.Mock
}

func (m *MockLocationService) UpdateDriverLocation(ctx context.Context, driverID string, lat, lon float64) error {
	args := m.Called(ctx, driverID, lat, lon)
	return args.Error(0)
}

func (m *MockLocationService) UpdateDriverLocations(ctx context.Context, locations []domain.DriverLocation) error {
	args := m.Called(ctx, locations)
	return args.Error(0)
}

func (m *MockLocationService) FindNearbyDrivers(ctx context.Context, lat, lon, radius float64) ([]*domain.DriverLocation, error) {
	args := m.Called(ctx, lat, lon, radius)
	return args.Get(0).([]*domain.DriverLocation), args.Error(1)
}

// MockTransport is a mock implementation of http.RoundTripper
type MockTransport struct {
	loginResponseBody         []byte
	nearbyDriversResponseBody []byte
}

func (m *MockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.URL.Path == "/api/v1/auth/login" {
		body = m.loginResponseBody
	} else if req.URL.Path == "/api/v1/locations/nearby" {
		body = m.nearbyDriversResponseBody
	} else {
		return nil, errors.New("unexpected URL")
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewBuffer(body)),
	}, nil
}

func TestFindNearestDriver(t *testing.T) {
	mockLocationService := new(MockLocationService)
	service := NewMatchingService(mockLocationService)

	// Mock the login response
	loginResponse := map[string]string{"token": "mockToken"}
	loginResponseBody, _ := json.Marshal(loginResponse)

	// Mock the nearby drivers response
	driverLocation := &domain.DriverLocation{
		ID:       "1",
		DriverID: "driver1",
		Location: domain.NewPoint(40.7128, -74.0060),
		Status:   "active",
	}
	nearbyDriversResponse := []*domain.DriverLocation{driverLocation}
	nearbyDriversResponseBody, _ := json.Marshal(nearbyDriversResponse)

	// Mock the HTTP client
	httpClient := &http.Client{
		Transport: &MockTransport{
			loginResponseBody:         loginResponseBody,
			nearbyDriversResponseBody: nearbyDriversResponseBody,
		},
	}

	// Replace the default HTTP client with the mock client
	oldClient := http.DefaultClient
	http.DefaultClient = httpClient
	defer func() { http.DefaultClient = oldClient }()

	// Call the method
	lat, lon, radius := 40.730610, -73.935242, 10.0
	driver, err := service.FindNearestDriver(context.Background(), lat, lon, radius)

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, driver)
	assert.Equal(t, "driver1", driver.DriverID)
}

func TestCalculateDistance(t *testing.T) {
	service := NewMatchingService(nil)

	lat1, lon1 := 40.7128, -74.0060
	lat2, lon2 := 34.0522, -118.2437
	expectedDistance := 3940.07 // Approximate distance in km

	distance := service.CalculateDistance(lat1, lon1, lat2, lon2)

	assert.InDelta(t, expectedDistance, distance, 1.0)
}
