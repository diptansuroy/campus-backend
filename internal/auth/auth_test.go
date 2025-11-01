package auth

import (
	"campus-backend/internal/users"
	"campus-backend/pkg/db"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var validate = validator.New()

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to test database")
	}
	
	// Auto migrate test models
	db.AutoMigrate(&users.User{})
	
	return db
}

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	hashed, err := HashPassword(password)
	
	assert.NoError(t, err)
	assert.NotEqual(t, password, hashed)
	assert.Len(t, hashed, 60) // bcrypt hash length
}

func TestCheckPasswordHash(t *testing.T) {
	password := "testpassword123"
	hashed, _ := HashPassword(password)
	
	// Test correct password
	assert.True(t, CheckPasswordHash(password, hashed))
	
	// Test incorrect password
	assert.False(t, CheckPasswordHash("wrongpassword", hashed))
}

func TestGenerateJWT(t *testing.T) {
	email := "test@example.com"
	role := "student"
	
	token, err := GenerateJWT(email, role)
	
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.Len(t, token, 150) // Approximate JWT length
}

func TestValidateStruct(t *testing.T) {
	// Test valid struct
	validReq := RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "password123",
		Role:     "student",
		Dept:     "Computer Science",
	}
	
	err := validate.ValidateStruct(validReq)
	assert.NoError(t, err)
	
	// Test invalid struct
	invalidReq := RegisterRequest{
		Name:     "J", // Too short
		Email:    "invalid-email", // Invalid email
		Password: "123", // Too short
		Role:     "invalid-role", // Invalid role
		Dept:     "", // Required field missing
	}
	
	err = validate.ValidateStruct(invalidReq)
	assert.Error(t, err)
}

func TestFormatValidationErrors(t *testing.T) {
	invalidReq := RegisterRequest{
		Name:     "J",
		Email:    "invalid-email",
		Password: "123",
		Role:     "invalid-role",
		Dept:     "",
	}
	
	err := validate.ValidateStruct(invalidReq)
	errors := validate.FormatValidationErrors(err)
	
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors, "Name")
	assert.Contains(t, errors, "Email")
	assert.Contains(t, errors, "Password")
	assert.Contains(t, errors, "Role")
	assert.Contains(t, errors, "Dept")
}

// Integration test for user registration
func TestUserRegistration(t *testing.T) {
	// Setup test database
	testDB := setupTestDB()
	db.DB = testDB
	
	// Test data
	req := RegisterRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
		Role:     "student",
		Dept:     "Computer Science",
	}
	
	// Validate request
	err := validate.ValidateStruct(req)
	assert.NoError(t, err)
	
	// Check if email already exists (should not exist)
	var existingUser users.User
	err = db.DB.Where("email = ?", req.Email).First(&existingUser).Error
	assert.Error(t, err) // Should error because user doesn't exist
	
	// Hash password
	hashedPassword, err := HashPassword(req.Password)
	assert.NoError(t, err)
	
	// Create user
	user := users.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     req.Role,
		Dept:     req.Dept,
		IsActive: true,
	}
	
	err = db.DB.Create(&user).Error
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)
	
	// Verify user was created
	var createdUser users.User
	err = db.DB.Where("email = ?", req.Email).First(&createdUser).Error
	assert.NoError(t, err)
	assert.Equal(t, req.Name, createdUser.Name)
	assert.Equal(t, req.Email, createdUser.Email)
	assert.Equal(t, req.Role, createdUser.Role)
	assert.Equal(t, req.Dept, createdUser.Dept)
	assert.True(t, createdUser.IsActive)
}
