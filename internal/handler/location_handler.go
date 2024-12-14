package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yusufatac/bitaksi-case-study/internal/domain"
	"github.com/yusufatac/bitaksi-case-study/internal/service"
)

type LocationHandler struct {
	locationService service.LocationService
}

func NewLocationHandler(locationService service.LocationService) *LocationHandler {
	return &LocationHandler{
		locationService: locationService,
	}
}

// UpdateLocation godoc
// @Summary Update driver location
// @Description Update a single driver's location using latitude and longitude
// @Tags locations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body UpdateLocationRequest true "Location update request"
// @Success 200 {object} Response "Location successfully updated"
// @Failure 400 {object} ErrorResponse "Invalid request parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /locations [post]
func (h *LocationHandler) UpdateLocation(c *gin.Context) {
	var req UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := h.locationService.UpdateDriverLocation(c, req.DriverID, req.Latitude, req.Longitude); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Message: "location updated successfully"})
}

// UpdateLocations godoc
// @Summary Update multiple driver locations
// @Description Update locations for multiple drivers in batch
// @Tags locations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body []domain.DriverLocation true "Batch location update request"
// @Success 200 {object} Response "Locations successfully updated"
// @Failure 400 {object} ErrorResponse "Invalid request parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /locations/batch [post]
func (h *LocationHandler) UpdateLocations(c *gin.Context) {
	var locations []domain.DriverLocation
	if err := c.ShouldBindJSON(&locations); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := h.locationService.UpdateDriverLocations(c, locations); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, Response{Message: "locations updated successfully"})
}

// FindNearbyDrivers godoc
// @Summary Find nearby drivers
// @Description Find drivers within a specified radius of a given location
// @Tags locations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body FindDriversRequest true "Find drivers request"
// @Success 200 {array} domain.DriverLocation "List of nearby drivers"
// @Failure 400 {object} ErrorResponse "Invalid request parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /locations/nearby [post]
func (h *LocationHandler) FindNearbyDrivers(c *gin.Context) {
	var req FindDriversRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	drivers, err := h.locationService.FindNearbyDrivers(c, req.Latitude, req.Longitude, req.Radius)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, drivers)
}

// Request/Response types
type UpdateLocationRequest struct {
	DriverID  string  `json:"driver_id" binding:"required"`
	Latitude  float64 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" binding:"required,min=-180,max=180"`
}

type FindDriversRequest struct {
	Latitude  float64 `json:"latitude" binding:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" binding:"required,min=-180,max=180"`
	Radius    float64 `json:"radius" binding:"required,gt=0"`
}

type Response struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
