package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/umahmood/haversine"
	"io/ioutil"
	"net/http"
	"sort"

	"github.com/yusufatac/bitaksi-case-study/internal/domain"
)

var (
	ErrNoDriversFound = errors.New("no drivers found within the specified radius")
)

type MatchingService interface {
	FindNearestDriver(ctx context.Context, lat, lon, radius float64) (*domain.DriverLocation, error)
	CalculateDistance(lat1, lon1, lat2, lon2 float64) float64
}

type matchingService struct {
	locationService LocationService
}

func NewMatchingService(locationService LocationService) MatchingService {
	return &matchingService{
		locationService: locationService,
	}
}

func (s *matchingService) FindNearestDriver(ctx context.Context, lat, lon, radius float64) (*domain.DriverLocation, error) {
	// Step 1: Login to get the JWT token
	loginURL := "http://driver-location-api:8080/api/v1/auth/login"
	loginReqBody, _ := json.Marshal(map[string]string{
		"username": "yusuf",
		"password": "secret",
	})

	loginResp, err := http.Post(loginURL, "application/json", bytes.NewBuffer(loginReqBody))
	if err != nil {
		return nil, err
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to login, status code: %d", loginResp.StatusCode)
	}

	var loginRespBody map[string]string
	if err := json.NewDecoder(loginResp.Body).Decode(&loginRespBody); err != nil {
		return nil, err
	}

	token, ok := loginRespBody["token"]
	if !ok {
		return nil, errors.New("token not found in login response")
	}

	// Step 2: Use the token to get nearby drivers
	url := "http://driver-location-api:8080/api/v1/locations/nearby"
	reqBody, _ := json.Marshal(map[string]interface{}{
		"latitude":  lat,
		"longitude": lon,
		"radius":    radius,
	})

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get nearby drivers, status code: %d", resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var drivers []*domain.DriverLocation
	if err := json.Unmarshal(body, &drivers); err != nil {
		return nil, err
	}

	if len(drivers) == 0 {
		return nil, ErrNoDriversFound
	}

	// Calculate distances and sort by distance
	type driverWithDistance struct {
		driver   *domain.DriverLocation
		distance float64
	}

	driversWithDistance := make([]driverWithDistance, len(drivers))
	for i, driver := range drivers {
		driverLat, driverLon := driver.Location.GetCoordinates()
		distance := s.CalculateDistance(lat, lon, driverLat, driverLon)
		driversWithDistance[i] = driverWithDistance{
			driver:   driver,
			distance: distance,
		}
	}

	// Sort by distance
	sort.Slice(driversWithDistance, func(i, j int) bool {
		return driversWithDistance[i].distance < driversWithDistance[j].distance
	})

	return driversWithDistance[0].driver, nil
}

// CalculateDistance calculates the distance between two points using the Haversine formula
func (s *matchingService) CalculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	c1 := haversine.Coord{Lat: lat1, Lon: lon1}
	c2 := haversine.Coord{Lat: lat2, Lon: lon2}
	_, km := haversine.Distance(c1, c2)

	return km
}
