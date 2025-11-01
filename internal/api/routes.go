package api

import (
	"campus-backend/internal/analytics"
	"campus-backend/internal/attendance"
	"campus-backend/internal/auth"
	"campus-backend/internal/leaves"
	"campus-backend/internal/notifications"
	"campus-backend/internal/users"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(r *gin.Engine) {
	// API group for version 1
	api := r.Group("/api/v1")

	// AUTH routes
	api.POST("/auth/register", auth.Register)
	api.POST("/auth/login", auth.Login)

	// USER routes
	api.GET("/users/me", auth.JWTAuthMiddleware(), users.MeHandler)
	api.GET("/users/", auth.JWTAuthMiddleware(), auth.RequireRole(users.RoleAdmin), users.ListUsers)
	api.GET("/admin/dashboard", auth.JWTAuthMiddleware(), auth.RequireRole(users.RoleAdmin), adminDashboardHandler)
	api.GET("/warden/dashboard", auth.JWTAuthMiddleware(), auth.RequireRole(users.RoleWarden), wardenDashboardHandler)
	api.GET("/faculty/dashboard", auth.JWTAuthMiddleware(), auth.RequireRole(users.RoleFaculty), facultyDashboardHandler)

	// LEAVES routes
	leavesGroup := api.Group("/leaves")
	{
		leavesGroup.POST("/apply", auth.JWTAuthMiddleware(), leaves.ApplyLeave)
		leavesGroup.GET("/", auth.JWTAuthMiddleware(), leaves.ListLeaves)
		leavesGroup.GET("/my", auth.JWTAuthMiddleware(), leaves.ListLeaves)
		leavesGroup.GET("/:id", auth.JWTAuthMiddleware(), leaves.GetLeaveDetails)
		leavesGroup.PUT("/:id/approve", auth.JWTAuthMiddleware(), leaves.ApproveRejectLeave)
		leavesGroup.PUT("/:id/reject", auth.JWTAuthMiddleware(), leaves.ApproveRejectLeave)
	}

	// ATTENDANCE routes
	attendanceGroup := api.Group("/attendance")
	{
		attendanceGroup.POST("/mark", auth.JWTAuthMiddleware(), auth.RequireRole(users.RoleFaculty), attendance.MarkAttendance)
		attendanceGroup.GET("/", auth.JWTAuthMiddleware(), attendance.ViewAttendance)
		attendanceGroup.GET("/stats", auth.JWTAuthMiddleware(), attendance.GetStats)
		attendanceGroup.GET("/department", auth.JWTAuthMiddleware(), attendance.GetDepartmentStats)
	}

	// ANALYTICS routes
	analyticsGroup := api.Group("/analytics")
	{
		analyticsGroup.GET("/summary", auth.JWTAuthMiddleware(), auth.RequireRole(users.RoleAdmin), analytics.GetSummary)
		analyticsGroup.GET("/leaves", auth.JWTAuthMiddleware(), auth.RequireRole(users.RoleAdmin), analytics.GetLeaveAnalytics)
		analyticsGroup.GET("/attendance", auth.JWTAuthMiddleware(), auth.RequireRole(users.RoleAdmin), analytics.GetAttendanceAnalytics)
	}

	// NOTIFICATIONS routes
	notificationsGroup := api.Group("/notifications")
	{
		notificationsGroup.GET("/", auth.JWTAuthMiddleware(), notifications.GetNotifications)
		notificationsGroup.GET("/unread-count", auth.JWTAuthMiddleware(), notifications.GetUnreadCount)
		notificationsGroup.PUT("/:id/read", auth.JWTAuthMiddleware(), notifications.MarkNotificationAsRead)
		notificationsGroup.PUT("/read-all", auth.JWTAuthMiddleware(), notifications.MarkAllNotificationsAsRead)
	}
}

// Dashboard handlers
func adminDashboardHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Welcome, Admin!"})
}

func wardenDashboardHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Welcome, Warden!"})
}

func facultyDashboardHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Welcome, Faculty!"})
}
