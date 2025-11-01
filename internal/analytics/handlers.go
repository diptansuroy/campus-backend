package analytics

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// DashboardStats struct - holds dashboard data
type DashboardStats struct {
	TotalStudents     int64   `json:"total_students"`
	TotalLeaves       int64   `json:"total_leaves"`
	PendingLeaves     int64   `json:"pending_leaves"`
	AverageAttendance float64 `json:"average_attendance"`
}

// GetSummary function - gets dashboard summary for admin
func GetSummary(c *gin.Context) {
	// Create service instance
	service := NewService()

	// Get dashboard stats
	stats, err := service.GetDashboardSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Send stats as JSON
	c.JSON(http.StatusOK, stats)
}

// GetLeaveAnalytics function - gets leave analytics for admin
func GetLeaveAnalytics(c *gin.Context) {
	// Create service instance
	service := NewService()

	// Get analytics data
	analytics, err := service.GetLeaveAnalytics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Send analytics as JSON
	c.JSON(http.StatusOK, analytics)
}

// GetAttendanceAnalytics function - gets attendance analytics for admin
func GetAttendanceAnalytics(c *gin.Context) {
	// Create service instance
	service := NewService()

	// Get analytics data
	analytics, err := service.GetAttendanceAnalytics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Send analytics as JSON
	c.JSON(http.StatusOK, analytics)
}

// AbsenteeRecord struct - holds absentee data
type AbsenteeRecord struct {
	StudentID   uint   `json:"student_id"`
	StudentName string `json:"student_name"`
	LeaveCount  int    `json:"leave_count"`
}
