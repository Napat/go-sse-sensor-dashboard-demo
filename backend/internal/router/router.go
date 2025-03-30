package router

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/Napat/go-sse-sensor-dashboard-demo/backend/internal/handler"
	"github.com/Napat/go-sse-sensor-dashboard-demo/backend/pkg/apierror"
	"github.com/Napat/go-sse-sensor-dashboard-demo/backend/pkg/config"
)

// IRouter คือ interface สำหรับจัดการ router
type IRouter interface {
	Setup(cfg *config.Config, log *zap.Logger) *echo.Echo
}

// Router เป็น implementation ของ IRouter
type Router struct{}

// NewRouter สร้าง instance ใหม่ของ Router
func NewRouter() *Router {
	return &Router{}
}

// SetupRouter สร้าง Echo instance และตั้งค่า middleware
func SetupRouter(cfg *config.Config, log *zap.Logger) *echo.Echo {
	e := echo.New()

	// จำกัดจำนวน concurrent requests
	store := middleware.NewRateLimiterMemoryStoreWithConfig(
		middleware.RateLimiterMemoryStoreConfig{
			Rate:      rate.Limit(cfg.MaxConnections),
			Burst:     int(float64(cfg.MaxConnections) * 1.5),
			ExpiresIn: time.Duration(cfg.WriteTimeout),
		},
	)

	e.Use(middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store:   store,
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return &echo.HTTPError{
				Code:     http.StatusTooManyRequests,
				Message:  "Too many requests",
				Internal: err,
			}
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return &echo.HTTPError{
				Code:     http.StatusTooManyRequests,
				Message:  "Too many requests",
				Internal: err,
			}
		},
	}))

	// ตั้งค่า CORS
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: strings.Split(cfg.CORSHosts, ","),
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))

	// ใช้ค่า security จาก config
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         cfg.Security.XSSProtection,
		ContentTypeNosniff:    cfg.Security.ContentTypeNosniff,
		XFrameOptions:         cfg.Security.XFrameOptions,
		HSTSMaxAge:            cfg.Security.HSTSMaxAge,
		ContentSecurityPolicy: cfg.Security.CSPPolicy,
	}))

	// Logger middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				log.Debug(v.URI,
					zap.Int("status", v.Status))
			} else {
				// ตรวจสอบว่าเป็น echo.HTTPError หรือไม่
				if echoErr, ok := v.Error.(*echo.HTTPError); ok && echoErr.Code < 500 {
					// สำหรับ HTTP error ทั่วไป (4xx) ไม่จำเป็นต้องแสดง stack trace
					log.Info("client request error",
						zap.String("URI", v.URI),
						zap.Int("status", v.Status),
						zap.String("error", echoErr.Error()))
				} else {
					// สำหรับ server error (5xx) ยังคงบันทึก stack trace
					log.Error("server request error",
						zap.String("URI", v.URI),
						zap.Int("status", v.Status),
						zap.Error(v.Error))
				}
			}
			return nil
		},
	}))

	// ให้ Echo ไม่แสดง Banner เมื่อเริ่มต้น server
	e.HideBanner = true
	e.HidePort = true

	// กำหนด custom error handler
	e.HTTPErrorHandler = customHTTPErrorHandler(log)

	// เรียกฟังก์ชัน setupRoutes
	if err := setupRoutes(e, cfg, log); err != nil {
		log.Fatal("Failed to setup routes", zap.Error(err))
	}

	return e
}

// customHTTPErrorHandler สร้าง HTTP error handler แบบกำหนดเอง
func customHTTPErrorHandler(log *zap.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		// ถ้าเป็น echo.HTTPError แปลงเป็น APIError
		if echoErr, ok := err.(*echo.HTTPError); ok {
			var message string
			switch m := echoErr.Message.(type) {
			case string:
				message = m
			case error:
				message = m.Error()
			default:
				message = "Unknown error"
			}

			apiErr := apierror.NewAPIError(
				"HTTP_ERROR",
				message,
				echoErr.Code,
			)
			// log.Error("HTTP error", zap.String("path", c.Path()), zap.Error(err))
			_ = c.JSON(apiErr.StatusCode(), apiErr)
			return
		}

		// แปลง error เป็น APIError
		apiErr := apierror.FromError(err)
		log.Error("API error", zap.String("path", c.Path()), zap.Error(err))

		// ส่งข้อมูลกลับไปยัง client
		if !c.Response().Committed {
			if c.Request().Method == http.MethodHead {
				_ = c.NoContent(apiErr.StatusCode())
			} else {
				_ = c.JSON(apiErr.StatusCode(), apiErr)
			}
		}
	}
}

// setupRoutes ตั้งค่า routes สำหรับแอปพลิเคชัน
func setupRoutes(e *echo.Echo, cfg *config.Config, log *zap.Logger) error {
	// Server static files จาก frontend
	e.Static("/", cfg.StaticPath)

	// ตั้งค่า API routes
	api := e.Group("/api")

	// สร้าง handler instances
	sensorHandler := handler.NewSensorHandler(log)

	// Sensor endpoints
	api.GET("/sensors/stream", sensorHandler.HandleSSE)
	api.GET("/sensors", sensorHandler.GetSensorData)
	api.GET("/sensors/:id", sensorHandler.GetSensorByID)

	// Environment endpoint
	api.GET("/environment", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"environment": cfg.Env,
			"production":  cfg.IsProduction(),
			"uat":         cfg.IsUAT(),
			"development": cfg.IsDevelopment(),
		})
	})

	// Health check endpoint
	e.GET("/health", func(c echo.Context) error {
		log.Debug("Health check requested")
		return c.JSON(http.StatusOK, map[string]string{
			"status": "ok",
			"env":    string(cfg.Env),
		})
	})

	return nil
}

// Setup ตั้งค่า Echo instance และ middleware (สำหรับ IRouter interface)
func (r *Router) Setup(cfg *config.Config, log *zap.Logger) *echo.Echo {
	return SetupRouter(cfg, log)
}
