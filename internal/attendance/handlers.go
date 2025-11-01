package attendance

import (
	"campus-backend/internal/users"
	"campus-backend/pkg/db"
	"campus-backend/pkg/validation"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type MarkAttendanceRequest struct {
	StudentID uint      `json:"student_id" binding:"required" validate:"required"`
	Date      time.Time `json:"date" binding:"required" validate:"required"`
	Present   bool      `json:"present" binding:"required"`
	Subject   *string   `json:"subject,omitempty" validate:"max=50"`
	Period    *string   `json:"period,omitempty" validate:"max=20"`
}

type AttendanceStats struct {
	StudentID            uint       `json:"student_id"`
	StudentName          string     `json:"student_name"`
	TotalDays            int        `json:"total_days"`
	PresentDays          int        `json:"present_days"`
	AbsentDays           int        `json:"absent_days"`
	AttendancePercentage float64    `json:"attendance_percentage"`
	LastAttendance       *time.Time `json:"last_attendance,omitempty"`
}

// MarkAttendance godoc
// @Summary Mark student attendance
// @Description Faculty marks attendance for a student
// @Tags Attendance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body MarkAttendanceRequest true "Attendance data"
// @Success 201 {object} map[string]interface{} "Attendance marked successfully"
// @Failure 400 {object} map[string]interface{} "Validation failed or attendance already marked"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "Student not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /attendance/mark [post]
func MarkAttendance(c *gin.Context) {
	var req MarkAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the struct
	if err := validation.ValidateStruct(req); err != nil {
		errors := validation.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": errors})
		return
	}

	markerIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	markerID := markerIDVal.(uint)

	// Check if student exists
	var student users.User
	if err := db.DB.First(&student, req.StudentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	// Check if attendance already exists for this date
	var existingAttendance Attendance
	err := db.DB.Where("student_id = ? AND date = ?", req.StudentID, req.Date.Truncate(24*time.Hour)).First(&existingAttendance).Error
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Attendance already marked for this date"})
		return
	}

	// Check if student has approved leave for this date
	var approvedLeave users.LeaveRequest
	err = db.DB.Where("student_id = ? AND status = ? AND start_date <= ? AND end_date >= ?",
		req.StudentID, "approved", req.Date.Truncate(24*time.Hour), req.Date.Truncate(24*time.Hour)).First(&approvedLeave).Error

	// If student has approved leave and is marked present, warn the faculty
	if err == nil && req.Present {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Student has approved leave for this date",
			"leave_details": gin.H{
				"leave_type": approvedLeave.LeaveType,
				"reason":     approvedLeave.Reason,
				"start_date": approvedLeave.StartDate,
				"end_date":   approvedLeave.EndDate,
			},
		})
		return
	}

	attendance := Attendance{
		StudentID: req.StudentID,
		Date:      req.Date.Truncate(24 * time.Hour),
		Present:   req.Present,
		MarkedBy:  markerID,
		Subject:   req.Subject,
		Period:    req.Period,
	}

	if err := db.DB.Create(&attendance).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark attendance"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Attendance marked successfully",
		"attendance": gin.H{
			"id":         attendance.ID,
			"student_id": attendance.StudentID,
			"date":       attendance.Date,
			"present":    attendance.Present,
			"subject":    attendance.Subject,
			"period":     attendance.Period,
			"marked_by":  attendance.MarkedBy,
			"created_at": attendance.CreatedAt,
		},
	})
}

func ViewAttendance(c *gin.Context) {
	roleVal, _ := c.Get("role")
	role := roleVal.(string)

	var studentID uint
	var err error

	// Determine which student's attendance to view
	if role == users.RoleStudent {
	studentIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
		studentID = studentIDVal.(uint)
	} else {
		// Faculty, Warden, or Admin can view any student's attendance
		studentIDParam := c.Query("student_id")
		if studentIDParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "student_id parameter is required"})
			return
		}
		// Parse studentID from string to uint
		// This is a simplified version - in production you'd want proper validation
		studentID = 1 // Placeholder - implement proper parsing
	}

	// Get query parameters for filtering
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	subject := c.Query("subject")

	var records []Attendance
	query := db.DB.Where("student_id = ?", studentID)

	if startDate != "" {
		if start, err := time.Parse("2006-01-02", startDate); err == nil {
			query = query.Where("date >= ?", start)
		}
	}
	if endDate != "" {
		if end, err := time.Parse("2006-01-02", endDate); err == nil {
			query = query.Where("date <= ?", end)
		}
	}
	if subject != "" {
		query = query.Where("subject = ?", subject)
	}

	err = query.Preload("Student").Preload("Marker").Order("date DESC").Find(&records).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve attendance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"attendance": records,
		"filters": gin.H{
			"start_date": startDate,
			"end_date":   endDate,
			"subject":    subject,
		},
	})
}

func GetStats(c *gin.Context) {
	roleVal, _ := c.Get("role")
	role := roleVal.(string)

	var studentID uint
	var err error

	// Determine which student's stats to get
	if role == users.RoleStudent {
		studentIDVal, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}
		studentID = studentIDVal.(uint)
	} else {
		// Faculty, Warden, or Admin can view any student's stats
		studentIDParam := c.Query("student_id")
		if studentIDParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "student_id parameter is required"})
			return
		}
		// Parse studentID from string to uint
		studentID = 1 // Placeholder - implement proper parsing
	}

	// Get student details
	var student users.User
	if err := db.DB.First(&student, studentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Student not found"})
		return
	}

	// Calculate attendance statistics
	var totalDays int64
	var presentDays int64
	var lastAttendance *time.Time

	// Count total attendance records
	err = db.DB.Model(&Attendance{}).Where("student_id = ?", studentID).Count(&totalDays).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate total days"})
		return
	}

	// Count present days
	err = db.DB.Model(&Attendance{}).Where("student_id = ? AND present = ?", studentID, true).Count(&presentDays).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate present days"})
		return
	}

	// Get last attendance date
	var lastRecord Attendance
	err = db.DB.Where("student_id = ?", studentID).Order("date DESC").First(&lastRecord).Error
	if err == nil {
		lastAttendance = &lastRecord.Date
	}

	// Calculate percentage
	var attendancePercentage float64
	if totalDays > 0 {
		attendancePercentage = float64(presentDays) / float64(totalDays) * 100
	}

	stats := AttendanceStats{
		StudentID:            studentID,
		StudentName:          student.Name,
		TotalDays:            int(totalDays),
		PresentDays:          int(presentDays),
		AbsentDays:           int(totalDays - presentDays),
		AttendancePercentage: attendancePercentage,
		LastAttendance:       lastAttendance,
	}

	c.JSON(http.StatusOK, stats)
}

func GetDepartmentStats(c *gin.Context) {
	roleVal, _ := c.Get("role")
	role := roleVal.(string)

	if role != users.RoleFaculty && role != users.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var dept string
	if role == users.RoleFaculty {
		userIDVal, _ := c.Get("userID")
		userID := userIDVal.(uint)
		var faculty users.User
		if err := db.DB.First(&faculty, userID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Faculty not found"})
			return
		}
		dept = faculty.Dept
	} else {
		dept = c.Query("department")
		if dept == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "department parameter is required"})
			return
		}
	}

	// Get all students in the department
	var students []users.User
	err := db.DB.Where("role = ? AND dept = ?", users.RoleStudent, dept).Find(&students).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get department students"})
		return
	}

	var departmentStats []AttendanceStats
	for _, student := range students {
		var totalDays int64
		var presentDays int64
		var lastAttendance *time.Time

		db.DB.Model(&Attendance{}).Where("student_id = ?", student.ID).Count(&totalDays)
		db.DB.Model(&Attendance{}).Where("student_id = ? AND present = ?", student.ID, true).Count(&presentDays)

		var lastRecord Attendance
		if err := db.DB.Where("student_id = ?", student.ID).Order("date DESC").First(&lastRecord).Error; err == nil {
			lastAttendance = &lastRecord.Date
		}

		var attendancePercentage float64
		if totalDays > 0 {
			attendancePercentage = float64(presentDays) / float64(totalDays) * 100
		}

		stats := AttendanceStats{
			StudentID:            student.ID,
			StudentName:          student.Name,
			TotalDays:            int(totalDays),
			PresentDays:          int(presentDays),
			AbsentDays:           int(totalDays - presentDays),
			AttendancePercentage: attendancePercentage,
			LastAttendance:       lastAttendance,
		}

		departmentStats = append(departmentStats, stats)
	}

	c.JSON(http.StatusOK, gin.H{
		"department":     dept,
		"stats":          departmentStats,
		"total_students": len(students),
	})
}
