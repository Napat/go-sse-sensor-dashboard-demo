package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestConnectionLimiter(t *testing.T) {
	// สร้าง Echo instance
	e := echo.New()

	tests := []struct {
		name           string
		maxConnections int64
		concurrency    int
		wantStatusCode int
	}{
		{
			name:           "Allow connection under limit",
			maxConnections: 5,
			concurrency:    1,
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "Reject connection over limit",
			maxConnections: 1,
			concurrency:    2,
			wantStatusCode: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// สร้าง connection limiter
			limiter := NewConnectionLimiter(tt.maxConnections)

			// สร้าง handler ที่จะถูกเรียกหลังจากผ่าน middleware
			handler := func(c echo.Context) error {
				return c.String(http.StatusOK, "OK")
			}

			// ส่งคำขอตามจำนวน concurrency
			statusCodes := make([]int, tt.concurrency)

			for i := 0; i < tt.concurrency; i++ {
				// สร้าง request และ response recorder
				req := httptest.NewRequest(http.MethodGet, "/", nil)
				rec := httptest.NewRecorder()
				c := e.NewContext(req, rec)

				// ดำเนินการด้วย middleware และ handler
				mw := limiter.Middleware()(handler)
				_ = mw(c)

				statusCodes[i] = rec.Code
			}

			// ตรวจสอบผลลัพธ์
			if tt.concurrency == 1 || tt.maxConnections >= int64(tt.concurrency) {
				// ทุกการเชื่อมต่อควรสำเร็จ
				for i, code := range statusCodes {
					if code != http.StatusOK {
						t.Errorf("Request %d expected status code %d, got %d", i, http.StatusOK, code)
					}
				}
			} else {
				// อย่างน้อยหนึ่งการเชื่อมต่อควรถูกปฏิเสธ
				rejected := false
				for _, code := range statusCodes {
					if code == http.StatusServiceUnavailable {
						rejected = true
						break
					}
				}

				if !rejected {
					t.Errorf("Expected at least one request to be rejected, but all succeeded")
				}
			}
		})
	}
}

func TestConnectionLimiterReleaseAndAcquire(t *testing.T) {
	// สร้าง connection limiter ที่อนุญาตเพียง 1 connection
	limiter := NewConnectionLimiter(1)

	// ทดสอบการขอและปล่อย token โดยตรง
	// ขอ token ครั้งแรก ควรสำเร็จ
	acquired1 := limiter.limiter.TryAcquire(1)
	if !acquired1 {
		t.Fatalf("Failed to acquire first token, expected success")
	}

	// ขอ token ครั้งที่สอง ควรล้มเหลวเพราะเกินขีดจำกัด
	acquired2 := limiter.limiter.TryAcquire(1)
	if acquired2 {
		// ปล่อย token ที่ขอสำเร็จ (ซึ่งไม่ควรเกิดขึ้น)
		limiter.limiter.Release(1)
		t.Fatalf("Acquired second token, expected failure")
	}

	// ปล่อย token แรก
	limiter.limiter.Release(1)

	// ขอ token ใหม่ หลังจากปล่อยแล้ว ควรสำเร็จ
	acquired3 := limiter.limiter.TryAcquire(1)
	if !acquired3 {
		t.Fatalf("Failed to acquire token after release, expected success")
	}

	// ต้องปล่อย token สุดท้ายเพื่อทำความสะอาด
	limiter.limiter.Release(1)
}
