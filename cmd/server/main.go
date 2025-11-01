// @title Campus Leave & Attendance Management System API
// @version 1.0
// @description A comprehensive backend system for campus leave and attendance management with role-based access control
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@campus.edu

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

package main

import (
	_ "campus-backend/docs" // Import docs for Swagger
	"campus-backend/internal/api"
	"campus-backend/internal/attendance"
	"campus-backend/internal/core"
	"campus-backend/internal/leaves"
	"campus-backend/internal/notifications"
	"campus-backend/internal/users"
	"campus-backend/pkg/db"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Load configuration
	config := core.LoadConfig()

	// Set Gin mode from config
	gin.SetMode(config.Server.GinMode)

	// Connect to database
	db.Connect()

	// Auto migrate tables - this creates tables automatically
	db.DB.AutoMigrate(&users.User{}, &leaves.LeaveRequest{}, &attendance.Attendance{}, &notifications.Notification{})

	// Create router
	r := gin.Default()

	// Setup all API routes using the api package
	api.SetupRoutes(r)

	// Add Swagger documentation route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server on configured port
	r.Run(":" + config.Server.Port)
}
