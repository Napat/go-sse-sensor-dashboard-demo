package apierror

import (
	"errors"
	"fmt"
	"net/http"
)

// Error codes สำหรับใช้ในระบบ
var (
	// Config errors
	ErrConfigNotFound     = errors.New("configuration file not found")
	ErrInvalidConfig      = errors.New("invalid configuration")
	ErrEnvironmentInvalid = errors.New("invalid environment")

	// Server errors
	ErrServerStartFailed = errors.New("failed to start server")
	ErrServerTimeout     = errors.New("server timeout")

	// Request errors
	ErrInvalidRequest   = errors.New("invalid request")
	ErrResourceNotFound = errors.New("resource not found")
	ErrUnauthorized     = errors.New("unauthorized access")
	ErrForbidden        = errors.New("forbidden access")

	// Data errors
	ErrDataNotFound = errors.New("data not found")
	ErrDataInvalid  = errors.New("invalid data")
	ErrDataConflict = errors.New("data conflict")
)

// APIError คือโครงสร้างสำหรับส่ง error กลับไปยัง client
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"-"` // HTTP status code, ไม่แสดงใน response
}

// Error ทำให้ APIError implement error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// StatusCode คืนค่า HTTP status code
func (e *APIError) StatusCode() int {
	return e.Status
}

// NewAPIError สร้าง APIError ใหม่
func NewAPIError(code string, message string, statusCode int) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Status:  statusCode,
	}
}

// FromError แปลง error ทั่วไปเป็น APIError
func FromError(err error) *APIError {
	switch {
	case errors.Is(err, ErrConfigNotFound), errors.Is(err, ErrInvalidConfig):
		return NewAPIError("CONFIG_ERROR", err.Error(), http.StatusInternalServerError)

	case errors.Is(err, ErrServerStartFailed), errors.Is(err, ErrServerTimeout):
		return NewAPIError("SERVER_ERROR", err.Error(), http.StatusInternalServerError)

	case errors.Is(err, ErrResourceNotFound), errors.Is(err, ErrDataNotFound):
		return NewAPIError("NOT_FOUND", err.Error(), http.StatusNotFound)

	case errors.Is(err, ErrUnauthorized):
		return NewAPIError("UNAUTHORIZED", err.Error(), http.StatusUnauthorized)

	case errors.Is(err, ErrForbidden):
		return NewAPIError("FORBIDDEN", err.Error(), http.StatusForbidden)

	case errors.Is(err, ErrInvalidRequest), errors.Is(err, ErrDataInvalid):
		return NewAPIError("INVALID_INPUT", err.Error(), http.StatusBadRequest)

	case errors.Is(err, ErrDataConflict):
		return NewAPIError("CONFLICT", err.Error(), http.StatusConflict)

	default:
		// ตรวจสอบว่าเป็น APIError อยู่แล้วหรือไม่
		var apiErr *APIError
		if errors.As(err, &apiErr) {
			return apiErr
		}

		// ถ้าไม่ใช่กรณีที่รู้จัก ให้คืนค่า internal server error
		return NewAPIError(
			"INTERNAL_SERVER_ERROR",
			"An unexpected error occurred",
			http.StatusInternalServerError,
		)
	}
}

// Wrap ห่อ error เดิมด้วยข้อความใหม่และคงการอ้างอิงไปยัง original error
func Wrap(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}

// HandleAPIError เป็น helper function สำหรับ handler ในการจัดการกับ error
func HandleAPIError(c interface{}, err error) error {
	// แปลง error เป็น APIError
	apiErr := FromError(err)

	// ตรวจสอบประเภทของ context
	type responder interface {
		JSON(int, interface{}) error
	}

	if responder, ok := c.(responder); ok {
		return responder.JSON(apiErr.StatusCode(), apiErr)
	}

	// ถ้า context ไม่ support JSON response
	return err
}

// IsError ตรวจสอบว่า error เป็นประเภทที่ระบุหรือไม่
func IsError(err error, target interface{}) bool {
	return errors.As(err, &target)
}
