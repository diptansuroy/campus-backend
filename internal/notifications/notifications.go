package notifications

import (
	"campus-backend/internal/users"
	"campus-backend/pkg/db"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	gorm.Model
	UserID    uint       `json:"user_id" gorm:"not null;index"`
	User      users.User `json:"user,omitempty" gorm:"foreignKey:UserID"`
	Title     string     `json:"title" gorm:"not null"`
	Message   string     `json:"message" gorm:"not null"`
	Type      string     `json:"type" gorm:"not null"` // leave_status, attendance, system
	IsRead    bool       `json:"is_read" gorm:"default:false"`
	RelatedID *uint      `json:"related_id,omitempty"` // ID of related leave request, etc.
	CreatedAt time.Time  `json:"created_at"`
}

type EmailService struct {
	// In a real implementation, you would use an email service like SendGrid, AWS SES, etc.
}

func NewEmailService() *EmailService {
	return &EmailService{}
}

func (e *EmailService) SendEmail(to, subject, body string) error {
	// Mock email sending - in production, integrate with actual email service
	log.Printf("Sending email to %s: %s - %s", to, subject, body)
	return nil
}

func CreateNotification(userID uint, title, message, notificationType string, relatedID *uint) error {
	notification := Notification{
		UserID:    userID,
		Title:     title,
		Message:   message,
		Type:      notificationType,
		RelatedID: relatedID,
	}

	return db.DB.Create(&notification).Error
}

func NotifyLeaveStatusChange(leaveRequest *users.LeaveRequest) error {
	var student users.User
	if err := db.DB.First(&student, leaveRequest.StudentID).Error; err != nil {
		return fmt.Errorf("failed to find student: %v", err)
	}

	var approver users.User
	if leaveRequest.ApprovedBy != nil {
		if err := db.DB.First(&approver, *leaveRequest.ApprovedBy).Error; err != nil {
			return fmt.Errorf("failed to find approver: %v", err)
		}
	}

	// Create notification for student
	title := fmt.Sprintf("Leave Request %s", leaveRequest.Status)
	message := fmt.Sprintf("Your leave request for %s (%s to %s) has been %s",
		leaveRequest.LeaveType,
		leaveRequest.StartDate.Format("2006-01-02"),
		leaveRequest.EndDate.Format("2006-01-02"),
		leaveRequest.Status)

	if leaveRequest.Remarks != nil {
		message += fmt.Sprintf(". Remarks: %s", *leaveRequest.Remarks)
	}

	err := CreateNotification(
		leaveRequest.StudentID,
		title,
		message,
		"leave_status",
		&leaveRequest.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to create notification: %v", err)
	}

	// Send email notification
	emailService := NewEmailService()
	emailSubject := fmt.Sprintf("Leave Request %s - Campus Management System", leaveRequest.Status)
	emailBody := fmt.Sprintf(`
Dear %s,

%s

Leave Details:
- Type: %s
- Reason: %s
- Start Date: %s
- End Date: %s
- Days: %d

%s

Best regards,
Campus Management System
`,
		student.Name,
		message,
		leaveRequest.LeaveType,
		leaveRequest.Reason,
		leaveRequest.StartDate.Format("2006-01-02"),
		leaveRequest.EndDate.Format("2006-01-02"),
		leaveRequest.Days,
		func() string {
			if leaveRequest.Remarks != nil {
				return fmt.Sprintf("Remarks: %s", *leaveRequest.Remarks)
			}
			return ""
		}(),
	)

	if err := emailService.SendEmail(student.Email, emailSubject, emailBody); err != nil {
		log.Printf("Failed to send email notification: %v", err)
	}

	return nil
}

func NotifyLeaveStartingTomorrow() error {
	tomorrow := time.Now().Add(24 * time.Hour).Truncate(24 * time.Hour)

	var leaves []users.LeaveRequest
	err := db.DB.Where("start_date = ? AND status = ?", tomorrow, "approved").Find(&leaves).Error
	if err != nil {
		return fmt.Errorf("failed to find leaves starting tomorrow: %v", err)
	}

	emailService := NewEmailService()

	for _, leave := range leaves {
		var student users.User
		if err := db.DB.First(&student, leave.StudentID).Error; err != nil {
			log.Printf("Failed to find student %d: %v", leave.StudentID, err)
			continue
		}

		// Create notification
		title := "Leave Starting Tomorrow"
		message := fmt.Sprintf("Your approved leave for %s starts tomorrow (%s). Please ensure all arrangements are in place.",
			leave.LeaveType, leave.StartDate.Format("2006-01-02"))

		err := CreateNotification(
			leave.StudentID,
			title,
			message,
			"leave_reminder",
			&leave.ID,
		)
		if err != nil {
			log.Printf("Failed to create notification for student %d: %v", leave.StudentID, err)
			continue
		}

		// Send email
		emailSubject := "Leave Starting Tomorrow - Reminder"
		emailBody := fmt.Sprintf(`
Dear %s,

%s

Leave Details:
- Type: %s
- Reason: %s
- Start Date: %s
- End Date: %s
- Days: %d

Please ensure all necessary arrangements are made before your leave begins.

Best regards,
Campus Management System
`,
			student.Name,
			message,
			leave.LeaveType,
			leave.Reason,
			leave.StartDate.Format("2006-01-02"),
			leave.EndDate.Format("2006-01-02"),
			leave.Days,
		)

		if err := emailService.SendEmail(student.Email, emailSubject, emailBody); err != nil {
			log.Printf("Failed to send reminder email to %s: %v", student.Email, err)
		}
	}

	return nil
}

func GetUserNotifications(userID uint, limit int) ([]Notification, error) {
	var notifications []Notification
	err := db.DB.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

func MarkNotificationAsReadDB(notificationID, userID uint) error {
	return db.DB.Model(&Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true).Error
}

func MarkAllNotificationsAsReadDB(userID uint) error {
	return db.DB.Model(&Notification{}).
		Where("user_id = ?", userID).
		Update("is_read", true).Error
}

func GetUnreadNotificationCount(userID uint) (int64, error) {
	var count int64
	err := db.DB.Model(&Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}
