package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/Napat/go-sse-sensor-dashboard-demo/backend/internal/router"
	"github.com/Napat/go-sse-sensor-dashboard-demo/backend/pkg/apierror"
	"github.com/Napat/go-sse-sensor-dashboard-demo/backend/pkg/config"
	"github.com/Napat/go-sse-sensor-dashboard-demo/backend/pkg/logger"
)

func main() {
	// เริ่มต้น logger ที่มี level เริ่มต้น (จะถูกปรับค่าหลังจากโหลด config)
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	// เริ่มต้น logger ชั่วคราวเพื่อใช้แสดงข้อความขณะโหลด config
	tempLog := logger.InitTempLogger(env)

	// พิมพ์ working directory ปัจจุบันเพื่อการตรวจสอบ
	currentDir, err := os.Getwd()
	if err != nil {
		tempLog.Fatal("Failed to get current directory", zap.Error(err))
	}
	tempLog.Info("Current working directory", zap.String("path", currentDir))

	// โหลด configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		tempLog.Fatal("Failed to load configuration", zap.Error(err))
	}

	// เริ่มต้น logger ถาวรด้วย log level จาก config
	log := logger.InitWithLevel(env, cfg.LogLevel)
	defer logger.Sync()

	log.Info("Starting server", zap.String("config", cfg.String()))

	// สร้าง root context พร้อมกับ cancel function
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ตั้งค่า router และ middleware
	r := router.NewRouter()
	e := r.Setup(cfg, log)

	// สร้าง server ด้วยค่า config
	server := &http.Server{
		Addr:           fmt.Sprintf(":%d", cfg.Port),
		Handler:        e,
		ReadTimeout:    cfg.ReadTimeout,
		WriteTimeout:   cfg.WriteTimeout,
		IdleTimeout:    cfg.IdleTimeout,
		MaxHeaderBytes: cfg.MaxHeaderBytes,
		BaseContext: func(listener net.Listener) context.Context {
			return rootCtx // ใช้ root context สำหรับทุก request
		},
	}

	// รัน server ในพร้อมกับการตรวจสอบสถานะ
	go func() {
		log.Info("Server started",
			zap.Int("port", cfg.Port),
			zap.String("env", string(cfg.Env)),
			zap.Int("maxConnections", cfg.MaxConnections))

		if err := e.StartServer(server); err != nil && err != http.ErrServerClosed {
			log.Error("Server error", zap.Error(err))
		}
	}()

	// ทำการ graceful shutdown
	waitForShutdown(e, log, cancel)
}

// waitForShutdown รอสัญญาณการปิดเซิร์ฟเวอร์และทำการปิดอย่างเรียบร้อย
func waitForShutdown(e *echo.Echo, log *zap.Logger, cancel context.CancelFunc) {
	// สร้าง channel สำหรับรับสัญญาณ
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// รอสัญญาณจาก OS
	<-quit
	log.Info("Shutdown signal received")

	// ยกเลิก root context
	cancel()

	// สร้าง context สำหรับการ timeout ของ shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Info("Shutting down server...")

	// ทำการปิด server อย่างถูกต้อง
	if err := e.Shutdown(ctx); err != nil {
		log.Error("Server shutdown error", zap.Error(apierror.Wrap(apierror.ErrServerTimeout, err.Error())))
		log.Fatal("Server forced to shutdown")
	}

	log.Info("Server gracefully stopped")
}
