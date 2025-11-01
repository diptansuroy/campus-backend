package leaves

import (
	"campus-backend/internal/notifications"
	"campus-backend/internal/users"
	"campus-backend/pkg/db"
	"campus-backend/pkg/validation"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type ApplyLeaveRequest struct {
	LeaveType string    `json:"leave_type" binding:"required" validate:"required,oneof=medical personal emergency academic"`
	Reason    string    `json:"reason" binding:"required" validate:"required,min=10,max=500"`
	StartDate time.Time `json:"start_date" binding:"required" validate:"required,future_date"`
	EndDate   time.Time `json:"end_date" binding:"required" validate:"required,date_range,leave_duration"`
}

type ApproveRejectRequest struct {
	Action  string  `json:"action" binding:"required" validate:"required,oneof=approve reject"`
	Remarks *string `json:"remarks" validate:"max=200"`
}

// ApplyLeave godoc
// @Summary Apply for leave
// @Description Student applies for leave with validation
// @Tags Leaves
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ApplyLeaveRequest true "Leave application data"
// @Success 201 {object} map[string]interface{} "Leave request submitted successfully"
// @Failure 400 {object} map[string]interface{} "Validation failed or overlapping leave"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /leaves/apply [post]
func ApplyLeave(c *gin.Context) {
	var input ApplyLeaveRequest

	// Get JSON data from request
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the data
	if err := validation.ValidateStruct(input); err != nil {
		errors := validation.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": errors})
		return
	}

	// Get student ID from JWT token
	studentIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	studentID := studentIDVal.(uint)

	// Get student details from database
	var student users.User
	if err := db.DB.First(&student, studentID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Student data not found"})
		return
	}

	// Check if student already has leave for same period
	var existingLeaves []LeaveRequest
	err := db.DB.Where("student_id = ? AND status IN (?) AND ((start_date <= ? AND end_date >= ?) OR (start_date <= ? AND end_date >= ?))",
		studentID, []string{"pending", "approved"}, input.StartDate, input.StartDate, input.EndDate, input.EndDate).Find(&existingLeaves).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check existing leaves"})
		return
	}

	// If overlapping leave exists, reject
	if len(existingLeaves) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You already have a leave request for this period"})
		return
	}

	// Calculate number of days
	days := int(input.EndDate.Sub(input.StartDate).Hours()/24) + 1

	// Create leave request
	leave := LeaveRequest{
		StudentID: studentID,
		LeaveType: input.LeaveType,
		Reason:    input.Reason,
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
		Status:    "pending", // Start as pending
		Dept:      student.Dept,
		Hostel:    student.Hostel,
		Days:      days,
	}

	// Save to database
	if err := db.DB.Create(&leave).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create leave request"})
		return
	}

	// Send success response
	c.JSON(http.StatusCreated, gin.H{
		"message": "Leave request submitted successfully",
		"leave_request": gin.H{
			"id":         leave.ID,
			"leave_type": leave.LeaveType,
			"reason":     leave.Reason,
			"start_date": leave.StartDate,
			"end_date":   leave.EndDate,
			"days":       leave.Days,
			"status":     leave.Status,
			"created_at": leave.CreatedAt,
		},
	})
}

// ListLeaves godoc
// @Summary List leave requests
// @Description Get list of leave requests based on user role
// @Tags Leaves
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (pending, approved, rejected)"
// @Param leave_type query string false "Filter by leave type"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "List of leave requests"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /leaves/ [get]
func ListLeaves(c *gin.Context) {
	roleVal, _ := c.Get("role")
	role := roleVal.(string)

	var leaves []LeaveRequest
	var err error

	// Get query parameters for filtering
	status := c.Query("status")
	leaveType := c.Query("leave_type")
	page := c.DefaultQuery("page", "1")
	limit := c.DefaultQuery("limit", "10")

	if role == users.RoleStudent {
		userIDVal, _ := c.Get("userID")
		userID := userIDVal.(uint)

		query := db.DB.Where("student_id = ?", userID)
		if status != "" {
			query = query.Where("status = ?", status)
		}
		if leaveType != "" {
			query = query.Where("leave_type = ?", leaveType)
		}

		err = query.Preload("Approver").Order("created_at DESC").Find(&leaves).Error
	} else if role == users.RoleWarden || role == users.RoleFaculty || role == users.RoleAdmin {
		// Filter leaves according to approval scope for warden and faculty
		if role == users.RoleWarden {
			userIDVal, _ := c.Get("userID")
			userID := userIDVal.(uint)
			var approver users.User
			if err := db.DB.First(&approver, userID).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
				return
			}

			query := db.DB.Where("hostel = ?", *approver.Hostel)
			if status != "" {
				query = query.Where("status = ?", status)
			} else {
				query = query.Where("status = ?", "pending") // Default to pending for wardens
			}
			if leaveType != "" {
				query = query.Where("leave_type = ?", leaveType)
			}

			err = query.Preload("Student").Preload("Approver").Order("created_at DESC").Find(&leaves).Error
		} else if role == users.RoleFaculty {
			userIDVal, _ := c.Get("userID")
			userID := userIDVal.(uint)
			var approver users.User
			if err := db.DB.First(&approver, userID).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
				return
			}

			query := db.DB.Where("dept = ?", approver.Dept)
			if status != "" {
				query = query.Where("status = ?", status)
			} else {
				query = query.Where("status = ?", "pending") // Default to pending for faculty
			}
			if leaveType != "" {
				query = query.Where("leave_type = ?", leaveType)
			}

			err = query.Preload("Student").Preload("Approver").Order("created_at DESC").Find(&leaves).Error
		} else {
			// Admin can see all leaves
			query := db.DB
			if status != "" {
				query = query.Where("status = ?", status)
			}
			if leaveType != "" {
				query = query.Where("leave_type = ?", leaveType)
			}

			err = query.Preload("Student").Preload("Approver").Order("created_at DESC").Find(&leaves).Error
		}
	} else {
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get leaves"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"leaves": leaves,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": len(leaves),
		},
	})
}

func GetLeaveDetails(c *gin.Context) {
	leaveID := c.Param("id")
	roleVal, _ := c.Get("role")
	role := roleVal.(string)

	var leave LeaveRequest
	if err := db.DB.Preload("Student").Preload("Approver").First(&leave, leaveID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Leave request not found"})
		return
	}

	// Check permissions
	if role == users.RoleStudent {
		userIDVal, _ := c.Get("userID")
		userID := userIDVal.(uint)
		if leave.StudentID != userID {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only view your own leave requests"})
			return
		}
	} else if role == users.RoleFaculty {
		userIDVal, _ := c.Get("userID")
		userID := userIDVal.(uint)
		var approver users.User
		if err := db.DB.First(&approver, userID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}
		if approver.Dept != leave.Dept {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only view leaves from your department"})
			return
		}
	} else if role == users.RoleWarden {
		userIDVal, _ := c.Get("userID")
		userID := userIDVal.(uint)
		var approver users.User
		if err := db.DB.First(&approver, userID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}
		if approver.Hostel == nil || leave.Hostel == nil || *approver.Hostel != *leave.Hostel {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only view leaves from your hostel"})
			return
		}
	}

	c.JSON(http.StatusOK, leave)
}

func ApproveRejectLeave(c *gin.Context) {
	leaveID := c.Param("id")

	var input ApproveRejectRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate the struct
	if err := validation.ValidateStruct(input); err != nil {
		errors := validation.FormatValidationErrors(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": errors})
		return
	}

	var leave LeaveRequest
	if err := db.DB.Preload("Student").First(&leave, leaveID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Leave request not found"})
		return
	}

	// Check if leave is already processed
	if leave.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Leave request has already been processed"})
		return
	}

	approverIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	approverID := approverIDVal.(uint)

	roleVal, _ := c.Get("role")
	role := roleVal.(string)

	// Role-based approval restrictions
	if role == users.RoleFaculty {
		// Faculty can only approve department leaves
		var approver users.User
		if err := db.DB.First(&approver, approverID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Approver not found"})
			return
		}
		if approver.Dept != leave.Dept {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only approve leaves from your department"})
			return
		}
	} else if role == users.RoleWarden {
		// Warden can only approve hostel leaves
		var approver users.User
		if err := db.DB.First(&approver, approverID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Approver not found"})
			return
		}
		if approver.Hostel == nil || leave.Hostel == nil || *approver.Hostel != *leave.Hostel {
			c.JSON(http.StatusForbidden, gin.H{"error": "You can only approve leaves from your hostel"})
			return
		}
	}

	// Update leave status
	switch input.Action {
	case "approve":
		leave.Status = "approved"
	case "reject":
		leave.Status = "rejected"
	}

	leave.ApprovedBy = &approverID
	leave.Remarks = input.Remarks

	if err := db.DB.Save(&leave).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update leave"})
		return
	}

	// TODO: Send notification to student about status change
	// Send notification to student about status change
	// Convert local LeaveRequest to users.LeaveRequest for notification
	userLeaveRequest := users.LeaveRequest{
		Model:      leave.Model,
		StudentID:  leave.StudentID,
		LeaveType:  leave.LeaveType,
		Reason:     leave.Reason,
		StartDate:  leave.StartDate,
		EndDate:    leave.EndDate,
		Status:     leave.Status,
		ApprovedBy: leave.ApprovedBy,
		Remarks:    leave.Remarks,
		Dept:       leave.Dept,
		Hostel:     leave.Hostel,
		Days:       leave.Days,
		CreatedAt:  leave.CreatedAt,
		UpdatedAt:  leave.UpdatedAt,
	}

	if err := notifications.NotifyLeaveStatusChange(&userLeaveRequest); err != nil {
		// Log error but don't fail the request
		// In production, you might want to use a proper logging system
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Leave request updated successfully",
		"leave_request": gin.H{
			"id":          leave.ID,
			"status":      leave.Status,
			"remarks":     leave.Remarks,
			"approved_by": leave.ApprovedBy,
			"updated_at":  leave.UpdatedAt,
		},
	})
}
