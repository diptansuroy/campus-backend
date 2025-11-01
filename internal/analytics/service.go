package analytics

type Service struct {
	repo *Repository
}

func NewService() *Service {
	return &Service{repo: NewRepository()}
}

func (s *Service) GetDashboardSummary() (*DashboardStats, error) {
	students, err := s.repo.GetStudentCount()
	if err != nil {
		return nil, err
	}

	total, pending, err := s.repo.GetLeaveStats()
	if err != nil {
		return nil, err
	}

	avg, err := s.repo.GetAttendanceAverage()
	if err != nil {
		return nil, err
	}

	return &DashboardStats{
		TotalStudents:     students,
		TotalLeaves:       total,
		PendingLeaves:     pending,
		AverageAttendance: avg,
	}, nil
}

func (s *Service) GetLeaveAnalytics() (map[string]interface{}, error) {
	// Monthly breakdown
	monthlyBreakdown, err := s.repo.GetMonthlyLeaveBreakdown()
	if err != nil {
		return nil, err
	}

	// Leave types distribution
	leaveTypes, err := s.repo.GetLeaveTypesDistribution()
	if err != nil {
		return nil, err
	}

	// Top absentees
	topAbsentees, err := s.repo.GetTopAbsentees()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"monthly_breakdown": monthlyBreakdown,
		"leave_types":       leaveTypes,
		"top_absentees":     topAbsentees,
	}, nil
}

func (s *Service) GetAttendanceAnalytics() (map[string]interface{}, error) {
	// Department-wise attendance
	deptWise, err := s.repo.GetDepartmentWiseAttendance()
	if err != nil {
		return nil, err
	}

	// Monthly trend
	monthlyTrend, err := s.repo.GetMonthlyAttendanceTrend()
	if err != nil {
		return nil, err
	}

	// Low attendance students
	lowAttendance, err := s.repo.GetLowAttendanceStudents()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"department_wise":         deptWise,
		"monthly_trend":           monthlyTrend,
		"low_attendance_students": lowAttendance,
	}, nil
}
