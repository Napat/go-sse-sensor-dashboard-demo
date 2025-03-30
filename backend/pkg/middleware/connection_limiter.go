package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/sync/semaphore"
)

// ConnectionLimiter คือ middleware สำหรับจำกัดจำนวน connection ที่เข้ามาพร้อมกัน
type ConnectionLimiter struct {
	// limiter ใช้สำหรับจำกัดจำนวน connection ด้วย semaphore
	limiter *semaphore.Weighted
}

// NewConnectionLimiter สร้าง instance ใหม่ของ ConnectionLimiter
func NewConnectionLimiter(maxConnections int64) *ConnectionLimiter {
	return &ConnectionLimiter{
		limiter: semaphore.NewWeighted(maxConnections),
	}
}

// Middleware สร้าง echo middleware function สำหรับจำกัดจำนวน connection
func (cl *ConnectionLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// พยายามขอ token จาก semaphore
			if !cl.limiter.TryAcquire(1) {
				return c.String(http.StatusServiceUnavailable, "Server is at maximum capacity. Please try again later.")
			}

			// ตั้งค่า callback ที่จะทำงานหลังจาก request เสร็จสิ้น
			c.Response().Before(func() {
				cl.limiter.Release(1)
			})

			// ดำเนินการต่อไปยัง handler ถัดไป
			return next(c)
		}
	}
}
