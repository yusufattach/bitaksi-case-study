package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yusufatac/bitaksi-case-study/internal/service"
)

type MatchingHandler struct {
	matchingService service.MatchingService
}

func NewMatchingHandler(matchingService service.MatchingService) *MatchingHandler {
	return &MatchingHandler{
		matchingService: matchingService,
	}
}

// FindNearestDriver godoc
// @Summary Find nearest driver
// @Description Find the nearest available driver within a specified radius
// @Tags matching
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body FindDriversRequest true "Find nearest driver request"
// @Success 200 {object} domain.DriverLocation "Nearest driver found"
// @Failure 400 {object} ErrorResponse "Invalid request parameters"
// @Failure 401 {object} ErrorResponse "Unauthorized"
// @Failure 404 {object} ErrorResponse "No drivers found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /match [post]
func (h *MatchingHandler) FindNearestDriver(c *gin.Context) {
	var req FindDriversRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	driver, err := h.matchingService.FindNearestDriver(c, req.Latitude, req.Longitude, req.Radius)
	if err != nil {
		if err == service.ErrNoDriversFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, driver)
}

type EstimateTimeRequest struct {
	PickupLatitude  float64 `json:"pickup_latitude" binding:"required,min=-90,max=90"`
	PickupLongitude float64 `json:"pickup_longitude" binding:"required,min=-180,max=180"`
	DriverLatitude  float64 `json:"driver_latitude" binding:"required,min=-90,max=90"`
	DriverLongitude float64 `json:"driver_longitude" binding:"required,min=-180,max=180"`
}

type EstimateTimeResponse struct {
	Distance      float64 `json:"distance_km"`
	EstimatedTime float64 `json:"estimated_time_minutes"`
}
