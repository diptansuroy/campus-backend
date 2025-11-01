package auth

import (
	"campus-backend/internal/users"
	"campus-backend/pkg/db"
	"campus-backend/pkg/validation"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Request structs for API
type RegisterRequest struct {
	Name      string  `json:"name" binding:"required" validate:"required,min=2,max=100"`
	Email     string  `json:"email" binding:"required" validate:"required,email"`
	Password  string  `json:"password" binding:"required" validate:"required,min=6"`
	Role      string  `json:"role" binding:"required" validate:"required,oneof=admin student faculty warden"`
	Dept      string  `json:"dept" binding:"required" validate:"required"`
	Hostel    *string `json:"hostel,omitempty"`
	Phone     *string `json:"phone,omitempty"`
	StudentID *string `json:"student_id,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required" validate:"required,email"`
	Password string `json:"password" binding:"required" validate:"required"`
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with the system
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "User registration data"
// @Success 201 {object} map[string]interface{} "User registered successfully"
// @Failure 400 {object} map[string]interface{} "Validation failed"
// @Failure 409 {object} map[string]interface{} "Email already registered"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/register [post]
func Register(c *gin.Context) {
	var req RegisterRequest

	// Get JSON data from request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the data
	if err := validation.ValidateStruct(req); err != nil {
		errors := validation.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": errors})
		return
	}

	// Check if email already exists
	var existingUser users.User
	if err := db.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Hash the password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create new user
	user := users.User{
		Name:      req.Name,
		Email:     req.Email,
		Password:  hashedPassword,
		Role:      req.Role,
		Dept:      req.Dept,
		Hostel:    req.Hostel,
		Phone:     req.Phone,
		StudentID: req.StudentID,
		IsActive:  true,
	}

	// Save to database
	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Don't send password back
	user.Password = ""

	// Send success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    user,
	})
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Failure 400 {object} map[string]interface{} "Validation failed"
// @Failure 401 {object} map[string]interface{} "Invalid credentials"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/login [post]
func Login(c *gin.Context) {
	var req LoginRequest

	// Get JSON data from request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the data
	if err := validation.ValidateStruct(req); err != nil {
		errors := validation.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": errors})
		return
	}

	// Find user by email
	var user users.User
	if err := db.DB.Where("email = ? AND is_active = ?", req.Email, true).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Check password
	if !CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate JWT token
	token, err := GenerateJWT(user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Update last login time
	now := time.Now()
	user.LastLogin = &now
	db.DB.Save(&user)

	// Don't send password back
	user.Password = ""

	// Send success response with token
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"user":    user,
	})
}

// List users by role - for admin use
func ListUsersByRole(c *gin.Context) {
	var users []users.User
	role := c.Query("role") // Get role from query parameter

	// Find users with specific role
	if err := db.DB.Where("role = ?", role).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Send users list
	c.JSON(http.StatusOK, users)
}
