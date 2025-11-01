package users

import (
	"campus-backend/pkg/db"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListUsers godoc
// @Summary List all users
// @Description Get list of all users (Admin only)
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param role query string false "Filter by role"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "List of users"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 403 {object} map[string]interface{} "Forbidden"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users/ [get]
func ListUsers(c *gin.Context) {
	var users []User
	var err error

	// Get query parameters for filtering
	role := c.Query("role")
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	// Build query
	query := db.DB
	if role != "" {
		query = query.Where("role = ?", role)
	}

	// Execute query
	err = query.Find(&users).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
		return
	}

	// Remove passwords from response
	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": len(users),
		},
	})
}

// MeHandler godoc
// @Summary Get current user profile
// @Description Get the profile of the currently authenticated user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "User profile"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users/me [get]
func MeHandler(c *gin.Context) {
	emailVal, ok := c.Get("email")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not in context"})
		return
	}
	email := emailVal.(string)

	var user User
	if err := db.DB.Where("email = ?", email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	user.Password = ""
	c.JSON(http.StatusOK, user)
}
