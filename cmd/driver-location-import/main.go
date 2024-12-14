package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/yusufatac/bitaksi-case-study/internal/domain"
)

type LoginResponse struct {
	Token string `json:"token"`
}

func login(username, password string) (string, error) {
	url := "http://localhost:8080/api/v1/auth/login"
	reqBody, _ := json.Marshal(map[string]string{
		"username": username,
		"password": password,
	})

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to login, status code: %d", resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var result LoginResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.Token, nil
}

func readCSV(filePath string, locationsChan chan<- domain.DriverLocation, wg *sync.WaitGroup) {
	defer wg.Done()

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV file: %v", err)
	}

	for i, record := range records {
		if i == 0 {
			// Skip header row
			continue
		}
		lat, err := strconv.ParseFloat(record[0], 64)
		if err != nil {
			log.Fatalf("Failed to parse latitude: %v", err)
		}
		long, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Fatalf("Failed to parse longitude: %v", err)
		}
		locationsChan <- domain.DriverLocation{
			DriverID:  uuid.New().String(),
			Location:  domain.NewPoint(lat, long),
			Status:    "active",
			Timestamp: time.Now(),
		}
	}
	close(locationsChan)
}

func updateDriverLocations(token string, locationsChan <-chan domain.DriverLocation, wg *sync.WaitGroup) {
	defer wg.Done()

	client := &http.Client{}
	url := "http://localhost:8080/api/v1/locations/batch"
	batchSize := 10
	var batch []domain.DriverLocation

	for location := range locationsChan {
		batch = append(batch, location)
		if len(batch) >= batchSize {
			if err := sendBatch(client, url, token, batch); err != nil {
				log.Fatalf("Failed to update locations: %v", err)
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		if err := sendBatch(client, url, token, batch); err != nil {
			log.Fatalf("Failed to update locations: %v", err)
		}
	}
}

func sendBatch(client *http.Client, url, token string, batch []domain.DriverLocation) error {
	reqBody, _ := json.Marshal(batch)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update locations, status code: %d", resp.StatusCode)
	}

	return nil
}

func main() {
	// Login to get JWT token
	token, err := login("yusuf", "secret")
	if err != nil {
		log.Fatalf("Failed to login: %v", err)
	}

	locationsChan := make(chan domain.DriverLocation, 100)
	var wg sync.WaitGroup

	// Read CSV file
	wg.Add(1)
	go readCSV("cmd/driver-location-import/Coordinates.csv", locationsChan, &wg)

	// Update driver locations
	wg.Add(1)
	go updateDriverLocations(token, locationsChan, &wg)

	wg.Wait()
	log.Println("Driver locations updated successfully")
}
