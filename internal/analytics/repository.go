package analytics

import (
	"campus-backend/internal/attendance"
	"campus-backend/internal/leaves"
	"campus-backend/internal/users"
	"campus-backend/pkg/db"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository() *Repository {
	return &Repository{db: db.DB}
}

func (r *Repository) GetStudentCount() (int64, error) {
	var count int64
	err := r.db.Model(&users.User{}).Where("role = ?", "student").Count(&count).Error
	return count, err
}

func (r *Repository) GetLeaveStats() (total int64, pending int64, err error) {
	err = r.db.Model(&leaves.LeaveRequest{}).Count(&total).Error
	if err != nil {
		return
	}
	err = r.db.Model(&leaves.LeaveRequest{}).Where("status = ?", "pending").Count(&pending).Error
	return
}

func (r *Repository) GetAttendanceAverage() (float64, error) {
	var result struct {
		Average float64
	}
	err := r.db.Model(&attendance.Attendance{}).
		Select("AVG(CASE WHEN present THEN 1 ELSE 0 END) as average").
		Scan(&result).Error
	return result.Average * 100, err
}

func (r *Repository) GetMonthlyLeaveBreakdown() (map[string]int, error) {
	var results []struct {
		Month string
		Count int
	}

	err := r.db.Model(&leaves.LeaveRequest{}).
		Select("DATE_TRUNC('month', created_at) as month, COUNT(*) as count").
		Group("DATE_TRUNC('month', created_at)").
		Order("month DESC").
		Limit(12).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	breakdown := make(map[string]int)
	for _, result := range results {
		month := result.Month[:7] // Get YYYY-MM format
		breakdown[month] = result.Count
	}

	return breakdown, nil
}

func (r *Repository) GetLeaveTypesDistribution() (map[string]int, error) {
	var results []struct {
		LeaveType string
		Count     int
	}

	err := r.db.Model(&leaves.LeaveRequest{}).
		Select("leave_type, COUNT(*) as count").
		Group("leave_type").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	distribution := make(map[string]int)
	for _, result := range results {
		distribution[result.LeaveType] = result.Count
	}

	return distribution, nil
}

func (r *Repository) GetTopAbsentees() ([]AbsenteeRecord, error) {
	var results []AbsenteeRecord

	err := r.db.Table("users").
		Select("users.id as student_id, users.name as student_name, COUNT(leave_requests.id) as leave_count").
		Joins("LEFT JOIN leave_requests ON users.id = leave_requests.student_id AND leave_requests.status = 'approved'").
		Where("users.role = ?", "student").
		Group("users.id, users.name").
		Order("leave_count DESC").
		Limit(10).
		Scan(&results).Error

	return results, err
}

func (r *Repository) GetDepartmentWiseAttendance() (map[string]float64, error) {
	var results []struct {
		Dept          string
		AvgAttendance float64
	}

	err := r.db.Table("users").
		Select("users.dept, AVG(CASE WHEN attendance.present THEN 1 ELSE 0 END) * 100 as avg_attendance").
		Joins("LEFT JOIN attendance ON users.id = attendance.student_id").
		Where("users.role = ?", "student").
		Group("users.dept").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	deptWise := make(map[string]float64)
	for _, result := range results {
		deptWise[result.Dept] = result.AvgAttendance
	}

	return deptWise, nil
}

func (r *Repository) GetMonthlyAttendanceTrend() (map[string]float64, error) {
	var results []struct {
		Month         string
		AvgAttendance float64
	}

	err := r.db.Table("attendance").
		Select("DATE_TRUNC('month', date) as month, AVG(CASE WHEN present THEN 1 ELSE 0 END) * 100 as avg_attendance").
		Group("DATE_TRUNC('month', date)").
		Order("month DESC").
		Limit(12).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	trend := make(map[string]float64)
	for _, result := range results {
		month := result.Month[:7] // Get YYYY-MM format
		trend[month] = result.AvgAttendance
	}

	return trend, nil
}

func (r *Repository) GetLowAttendanceStudents() ([]AbsenteeRecord, error) {
	var results []AbsenteeRecord

	err := r.db.Table("users").
		Select("users.id as student_id, users.name as student_name, (COUNT(CASE WHEN attendance.present THEN 1 END) * 100.0 / COUNT(attendance.id)) as leave_count").
		Joins("LEFT JOIN attendance ON users.id = attendance.student_id").
		Where("users.role = ?", "student").
		Group("users.id, users.name").
		Having("(COUNT(CASE WHEN attendance.present THEN 1 END) * 100.0 / COUNT(attendance.id)) < 75").
		Order("leave_count ASC").
		Limit(10).
		Scan(&results).Error

	return results, err
}
