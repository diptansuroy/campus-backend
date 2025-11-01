package attendance

import (
	"time"

	"gorm.io/gorm"
)

// Attendance represents attendance records
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

// User represents a user (imported from users package)
type User struct {
	gorm.Model
	Name      string     `json:"name" gorm:"not null" validate:"required,min=2,max=100"`
	Email     string     `json:"email" gorm:"uniqueIndex;not null" validate:"required,email"`
	Password  string     `json:"-" gorm:"not null" validate:"required,min=6"`
	Role      string     `json:"role" gorm:"not null" validate:"required,oneof=admin student faculty warden"`
	Dept      string     `json:"dept" gorm:"not null" validate:"required"`
	Hostel    *string    `json:"hostel,omitempty"`
	Phone     *string    `json:"phone,omitempty"`
	StudentID *string    `json:"student_id,omitempty" gorm:"uniqueIndex"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	LastLogin *time.Time `json:"last_login,omitempty"`
}
