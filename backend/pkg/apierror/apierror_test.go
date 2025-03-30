package apierror_test

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Napat/go-sse-sensor-dashboard-demo/backend/pkg/apierror"
)

// MockResponder ใช้สำหรับทดสอบ HandleAPIError
type MockResponder struct {
	StatusCode int
	Response   interface{}
	Error      error
}

// JSON จำลองการตอบกลับ JSON
func (m *MockResponder) JSON(code int, response interface{}) error {
	m.StatusCode = code
	m.Response = response
	return m.Error
}

func TestNewAPIError(t *testing.T) {
	// ทดสอบการสร้าง APIError ใหม่
	apiErr := apierror.NewAPIError("TEST_CODE", "Test message", http.StatusBadRequest)

	assert.Equal(t, "TEST_CODE", apiErr.Code)
	assert.Equal(t, "Test message", apiErr.Message)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode())
	assert.Equal(t, "TEST_CODE: Test message", apiErr.Error())
}

func TestFromError(t *testing.T) {
	// ทดสอบการแปลง error เป็น APIError
	tests := []struct {
		name           string
		err            error
		expectedCode   string
		expectedStatus int
	}{
		{
			name:           "config_error",
			err:            apierror.ErrConfigNotFound,
			expectedCode:   "CONFIG_ERROR",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "server_error",
			err:            apierror.ErrServerTimeout,
			expectedCode:   "SERVER_ERROR",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "not_found_error",
			err:            apierror.ErrDataNotFound,
			expectedCode:   "NOT_FOUND",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "unauthorized_error",
			err:            apierror.ErrUnauthorized,
			expectedCode:   "UNAUTHORIZED",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "forbidden_error",
			err:            apierror.ErrForbidden,
			expectedCode:   "FORBIDDEN",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "invalid_request_error",
			err:            apierror.ErrInvalidRequest,
			expectedCode:   "INVALID_INPUT",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "data_conflict_error",
			err:            apierror.ErrDataConflict,
			expectedCode:   "CONFLICT",
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "unknown_error",
			err:            errors.New("unknown error"),
			expectedCode:   "INTERNAL_SERVER_ERROR",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "wrapped_error",
			err:            fmt.Errorf("wrapper: %w", apierror.ErrDataNotFound),
			expectedCode:   "NOT_FOUND",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			apiErr := apierror.FromError(tc.err)

			assert.Equal(t, tc.expectedCode, apiErr.Code)
			assert.Equal(t, tc.expectedStatus, apiErr.StatusCode())
		})
	}
}

func TestWrap(t *testing.T) {
	// ทดสอบการห่อ error
	originalErr := errors.New("original error")
	wrappedErr := apierror.Wrap(originalErr, "wrapped message")

	// ตรวจสอบว่า error ยังสามารถตรวจจับได้ด้วย errors.Is
	assert.True(t, errors.Is(wrappedErr, originalErr))

	// ตรวจสอบข้อความ error
	assert.Contains(t, wrappedErr.Error(), "wrapped message")
	assert.Contains(t, wrappedErr.Error(), "original error")
}

func TestHandleAPIError(t *testing.T) {
	// ทดสอบการแปลง error เป็น APIError และส่งกลับไปยัง client
	tests := []struct {
		name         string
		err          error
		responderErr error
		expectedCode int
	}{
		{
			name:         "success_case",
			err:          apierror.ErrDataNotFound,
			responderErr: nil,
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "responder_error",
			err:          apierror.ErrInvalidRequest,
			responderErr: errors.New("responder error"),
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			responder := &MockResponder{Error: tc.responderErr}

			result := apierror.HandleAPIError(responder, tc.err)

			// ตรวจสอบ status code ที่ถูกส่งไปยัง responder
			assert.Equal(t, tc.expectedCode, responder.StatusCode)

			// ตรวจสอบ error ที่ถูกส่งกลับ
			if tc.responderErr != nil {
				assert.Equal(t, tc.responderErr, result)
			} else {
				assert.Nil(t, result)
			}

			// ตรวจสอบว่า response เป็น APIError
			apiErr, ok := responder.Response.(*apierror.APIError)
			assert.True(t, ok)
			require.NotNil(t, apiErr)
		})
	}
}

func TestIsError(t *testing.T) {
	// ทดสอบการตรวจสอบประเภทของ error
	var apiErr *apierror.APIError

	// สร้าง APIError
	originalErr := apierror.NewAPIError("TEST", "test message", http.StatusOK)

	// ห่อ error
	wrappedErr := fmt.Errorf("wrapped: %w", originalErr)

	// ทดสอบการตรวจจับ
	assert.True(t, apierror.IsError(wrappedErr, &apiErr))
	assert.Equal(t, "TEST", apiErr.Code)
}

func TestAPIErrorCases(t *testing.T) {
	// ทดสอบการ chain error
	err1 := apierror.ErrDataNotFound
	err2 := apierror.Wrap(err1, "layer 1")
	err3 := apierror.Wrap(err2, "layer 2")

	// ตรวจสอบว่ายังสามารถค้นหา original error ได้
	assert.True(t, errors.Is(err3, apierror.ErrDataNotFound))

	// ทดสอบการแปลงเป็น APIError
	apiErr := apierror.FromError(err3)
	assert.Equal(t, "NOT_FOUND", apiErr.Code)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode())
}
