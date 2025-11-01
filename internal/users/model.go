package users

import (
	"time"

	"gorm.io/gorm"
)

// User struct - represents a user in the system
type User struct {
	gorm.Model
	Name      string     `json:"name" gorm:"not null" validate:"required,min=2,max=100"`
	Email     string     `json:"email" gorm:"uniqueIndex;not null" validate:"required,email"`
	Password  string     `json:"-" gorm:"not null" validate:"required,min=6"` // Don't show password in JSON
	Role      string     `json:"role" gorm:"not null" validate:"required,oneof=admin student faculty warden"`
	Dept      string     `json:"dept" gorm:"not null" validate:"required"`
	Hostel    *string    `json:"hostel,omitempty"`
	Phone     *string    `json:"phone,omitempty"`
	StudentID *string    `json:"student_id,omitempty" gorm:"uniqueIndex"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	LastLogin *time.Time `json:"last_login,omitempty"`

	// Relationships - these connect to other tables
	LeaveRequests []LeaveRequest `json:"leave_requests,omitempty" gorm:"foreignKey:StudentID"`
	Attendance    []Attendance   `json:"attendance,omitempty" gorm:"foreignKey:StudentID"`
}

// LeaveRequest struct - represents a leave request
type LeaveRequest struct {
	gorm.Model
	StudentID  uint      `json:"student_id" gorm:"not null;index"`
	Student    User      `json:"student,omitempty" gorm:"foreignKey:StudentID"`
	LeaveType  string    `json:"leave_type" gorm:"not null" validate:"required,oneof=medical personal emergency academic"`
	Reason     string    `json:"reason" gorm:"not null" validate:"required,min=10,max=500"`
	StartDate  time.Time `json:"start_date" gorm:"not null" validate:"required"`
	EndDate    time.Time `json:"end_date" gorm:"not null" validate:"required"`
	Status     string    `json:"status" gorm:"not null;default:pending" validate:"oneof=pending approved rejected"`
	ApprovedBy *uint     `json:"approved_by,omitempty" gorm:"index"`
	Approver   *User     `json:"approver,omitempty" gorm:"foreignKey:ApprovedBy"`
	Remarks    *string   `json:"remarks,omitempty" validate:"max=200"`
	Dept       string    `json:"dept" gorm:"not null"`
	Hostel     *string   `json:"hostel,omitempty"`
	Days       int       `json:"days" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Attendance struct - represents attendance records
type Attendance struct {
	gorm.Model
	StudentID uint      `json:"student_id" gorm:"not null;index"`
	Student   User      `json:"student,omitempty" gorm:"foreignKey:StudentID"`
	Date      time.Time `json:"date" gorm:"not null;index"`
	Present   bool      `json:"present" gorm:"not null"`
	MarkedBy  uint      `json:"marked_by" gorm:"not null"`
	Marker    User      `json:"marker,omitempty" gorm:"foreignKey:MarkedBy"`
	Subject   *string   `json:"subject,omitempty"`
	Period    *string   `json:"period,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
