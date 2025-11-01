# 🏫 Campus Leave & Attendance Management System

A comprehensive backend system designed to digitize and automate the leave and attendance workflow for universities and hostels. This system eliminates manual approvals, paper forms, and fragmented data — enabling students, wardens, and faculty to interact through a structured, auditable process.

## 🎯 Features

- 🔑 **JWT-based Authentication** with role-based authorization (Admin, Faculty, Warden, Student)
- 📝 **CRUD APIs** for users, leave requests, and attendance tracking
- ✅ **Approval/Rejection workflows** with remarks and status updates
- 📢 **Real-time Notifications** for leave status changes
- 📊 **Comprehensive Analytics** for attendance and absentee trends
- 🐳 **Docker Support** for easy deployment
- 📚 **API Documentation** with Swagger integration

## 🏗️ Architecture

```
/cmd
  /server           → main.go entry point
/internal
  /api              → route handlers
  /auth             → JWT logic & middleware
  /users            → user models & endpoints
  /leaves           → leave CRUD & approval flow
  /attendance       → attendance management
  /notifications    → async notification jobs
  /analytics        → data aggregation & reporting
/pkg
  /db               → database setup (GORM)
  /validation       → input validation utilities
```

## API Endpoints


| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `POST` | `/api/v1/auth/register` | Register a new user | No |
| `POST` | `/api/v1/auth/login` | Authenticate user | No |

### Users

| Method | Endpoint | Description | Auth Required | Role Required |
|--------|----------|-------------|---------------|---------------|
| `GET` | `/api/v1/users/me` | Get current user profile | Yes | Any |

### Leave Management

| Method | Endpoint | Description | Auth Required | Role Required |
|--------|----------|-------------|---------------|---------------|
| `POST` | `/api/v1/leaves/apply` | Submit new leave request | Yes | Student |
| `GET` | `/api/v1/leaves/` | List leave requests | Yes | Any |
| `GET` | `/api/v1/leaves/:id` | Get leave request details | Yes | Any |
| `PUT` | `/api/v1/leaves/:id/approve` | Approve leave request | Yes | Faculty/Warden |
| `PUT` | `/api/v1/leaves/:id/reject` | Reject leave request | Yes | Faculty/Warden |

### Attendance

| Method | Endpoint | Description | Auth Required | Role Required |
|--------|----------|-------------|---------------|---------------|
| `POST` | `/api/v1/attendance/mark` | Mark student attendance | Yes | Faculty |
| `GET` | `/api/v1/attendance/` | View attendance records | Yes | Any |
| `GET` | `/api/v1/attendance/stats` | Get attendance statistics | Yes | Any |
| `GET` | `/api/v1/attendance/department` | Get department-wise stats | Yes | Faculty/Admin |

### Analytics (Admin Only)

| Method | Endpoint | Description | Auth Required | Role Required |
|--------|----------|-------------|---------------|---------------|
| `GET` | `/api/v1/analytics/summary` | Dashboard summary | Yes | Admin |
| `GET` | `/api/v1/analytics/leaves` | Leave analytics | Yes | Admin |
| `GET` | `/api/v1/analytics/attendance` | Attendance analytics | Yes | Admin |

### Notifications

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/api/v1/notifications/` | Get user notifications | Yes |
| `GET` | `/api/v1/notifications/unread-count` | Get unread count | Yes |
| `PUT` | `/api/v1/notifications/:id/read` | Mark notification as read | Yes |
| `PUT` | `/api/v1/notifications/read-all` | Mark all as read | Yes |

## User Roles & Permissions

### Student
- Apply for leave requests
- View own leave requests and status
- View own attendance records
- Receive notifications about leave status

### Faculty
- Approve/reject department leave requests
- Mark student attendance
- View department attendance statistics
- View department leave requests

### Warden
- Approve/reject hostel-related leave requests
- View hostel attendance statistics
- Track frequent absentees

### Admin
- Full system access
- View all analytics and reports
- Manage users
- Monitor system-wide patterns


Enjoy!