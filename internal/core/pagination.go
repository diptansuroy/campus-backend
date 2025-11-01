package core

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Pagination struct for paginated responses
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// PaginationParams extracts pagination parameters from request
func PaginationParams(c *gin.Context) (page, limit int) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, _ = strconv.Atoi(pageStr)
	limit, _ = strconv.Atoi(limitStr)

	// Set defaults and limits
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	return page, limit
}

// CalculatePagination calculates pagination metadata
func CalculatePagination(page, limit int, total int64) Pagination {
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return Pagination{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// PaginatedResponse creates a paginated JSON response
func PaginatedResponse(c *gin.Context, data interface{}, pagination Pagination) {
	c.JSON(http.StatusOK, gin.H{
		"data":       data,
		"pagination": pagination,
	})
}

// ErrorResponse creates a standardized error response
func ErrorResponse(c *gin.Context, statusCode int, message string, details interface{}) {
	response := gin.H{"error": message}
	if details != nil {
		response["details"] = details
	}
	c.JSON(statusCode, response)
}

// SuccessResponse creates a standardized success response
func SuccessResponse(c *gin.Context, message string, data interface{}) {
	response := gin.H{"message": message}
	if data != nil {
		response["data"] = data
	}
	c.JSON(http.StatusOK, response)
}
