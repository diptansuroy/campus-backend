package validation

import (
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	
	// Register custom validators
	validate.RegisterValidation("date_range", validateDateRange)
	validate.RegisterValidation("future_date", validateFutureDate)
	validate.RegisterValidation("leave_duration", validateLeaveDuration)
}

// ValidateStruct validates a struct using the validator
func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

// validateDateRange ensures end date is after start date
func validateDateRange(fl validator.FieldLevel) bool {
	startDate := fl.Parent().FieldByName("StartDate")
	endDate := fl.Field()
	
	if !startDate.IsValid() || !endDate.IsValid() {
		return false
	}
	
	start, ok1 := startDate.Interface().(time.Time)
	end, ok2 := endDate.Interface().(time.Time)
	
	if !ok1 || !ok2 {
		return false
	}
	
	return end.After(start)
}

// validateFutureDate ensures the date is not in the past
func validateFutureDate(fl validator.FieldLevel) bool {
	date, ok := fl.Field().Interface().(time.Time)
	if !ok {
		return false
	}
	
	// Allow today's date but not past dates
	return !date.Before(time.Now().Truncate(24 * time.Hour))
}

// validateLeaveDuration ensures leave duration is reasonable (max 30 days)
func validateLeaveDuration(fl validator.FieldLevel) bool {
	startDate := fl.Parent().FieldByName("StartDate")
	endDate := fl.Field()
	
	if !startDate.IsValid() || !endDate.IsValid() {
		return false
	}
	
	start, ok1 := startDate.Interface().(time.Time)
	end, ok2 := endDate.Interface().(time.Time)
	
	if !ok1 || !ok2 {
		return false
	}
	
	duration := end.Sub(start)
	return duration <= 30*24*time.Hour && duration >= 0
}

// FormatValidationErrors formats validation errors into a readable format
func FormatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)
	
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			tag := e.Tag()
			
			switch tag {
			case "required":
				errors[field] = fmt.Sprintf("%s is required", field)
			case "email":
				errors[field] = fmt.Sprintf("%s must be a valid email address", field)
			case "min":
				errors[field] = fmt.Sprintf("%s must be at least %s characters long", field, e.Param())
			case "max":
				errors[field] = fmt.Sprintf("%s must be at most %s characters long", field, e.Param())
			case "oneof":
				errors[field] = fmt.Sprintf("%s must be one of: %s", field, e.Param())
			case "date_range":
				errors[field] = "End date must be after start date"
			case "future_date":
				errors[field] = "Date cannot be in the past"
			case "leave_duration":
				errors[field] = "Leave duration cannot exceed 30 days"
			default:
				errors[field] = fmt.Sprintf("%s is invalid", field)
			}
		}
	}
	
	return errors
}
