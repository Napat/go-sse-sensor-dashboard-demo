package handler

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/Napat/go-sse-sensor-dashboard-demo/backend/internal/service"
	"github.com/Napat/go-sse-sensor-dashboard-demo/backend/pkg/apierror"
)

// ISensorHandler คือ interface สำหรับ handler ที่จัดการเกี่ยวกับ sensor
type ISensorHandler interface {
	// HandleSSE จัดการ Server-Sent Events
	HandleSSE(c echo.Context) error

	// GetSensorData คืนค่าข้อมูล sensor ทั้งหมด
	GetSensorData(c echo.Context) error

	// GetSensorByID คืนค่าข้อมูล sensor ตาม ID
	GetSensorByID(c echo.Context) error
}

// SensorHandler จัดการเกี่ยวกับ handler ของ sensor API
type SensorHandler struct {
	// dependency ต่างๆ
	sensorService service.ISensorService
	logger        *zap.Logger
}

// NewSensorHandler สร้าง instance ใหม่ของ SensorHandler
func NewSensorHandler(logger *zap.Logger) *SensorHandler {
	return &SensorHandler{
		sensorService: service.GetSensorService(logger),
		logger:        logger,
	}
}

// HandleSSE จัดการกับ Server-Sent Events
func (h *SensorHandler) HandleSSE(c echo.Context) error {
	// ตั้งค่า header สำหรับ SSE
	c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().WriteHeader(http.StatusOK)

	// สร้าง channel เพื่อรับสัญญาณการปิดการเชื่อมต่อ
	done := make(chan bool)

	// ติดตามการตัดการเชื่อมต่อของ client
	go func() {
		<-c.Request().Context().Done()
		close(done)
	}()

	// ตรวจสอบ Last-Event-ID จาก request header (ถ้ามี)
	lastEventID := c.Request().Header.Get("Last-Event-ID")
	var lastTimestamp int64

	if lastEventID != "" {
		// แปลง string เป็น int64
		var err error
		lastTimestamp, err = strconv.ParseInt(lastEventID, 10, 64)
		if err != nil {
			h.logger.Warn("Invalid Last-Event-ID received",
				zap.String("last_event_id", lastEventID),
				zap.Error(err))
		} else {
			h.logger.Info("Reconnection with Last-Event-ID",
				zap.Int64("last_event_id", lastTimestamp),
				zap.String("client_ip", c.RealIP()))
		}
	}

	// บันทึก log การเชื่อมต่อ
	h.logger.Info("Client connected to SSE",
		zap.String("client_ip", c.RealIP()))

	// Flush buffer เพื่อให้ส่งข้อมูลเริ่มต้นได้ทันที
	c.Response().Flush()

	// อ่าน hostname ของ container
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
		h.logger.Warn("Unable to get hostname", zap.Error(err))
	}

	// ส่งข้อมูลเซ็นเซอร์เริ่มต้น
	initialData, err := h.sensorService.GetAllSensors()
	if err != nil {
		h.logger.Error("Failed to get initial sensor data",
			zap.Error(err))
		return apierror.HandleAPIError(c, apierror.Wrap(apierror.ErrDataNotFound, "failed to get sensor data"))
	}

	// ใช้ timestamp ปัจจุบันเป็น ID
	currentTimestamp := time.Now().Unix()

	// เพิ่ม hostname เข้าไปใน response
	initialDataWithHostname := fmt.Sprintf(`{"server_id":"%s","data":%s}`, hostname, initialData)

	// ส่งข้อมูลเริ่มต้นไปยัง client พร้อม ID
	fmt.Fprintf(c.Response(), "id: %d\nevent: message\ndata: %s\n\n", currentTimestamp, initialDataWithHostname)
	c.Response().Flush()

	// ตั้ง ticker สำหรับส่งข้อมูลทุก 2 วินาที
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// ส่ง ping ทุก 30 วินาที เพื่อรักษาการเชื่อมต่อ
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	// รับและส่งข้อมูลเมื่อมีการอัพเดท
	for {
		select {
		case <-done:
			h.logger.Info("Client disconnected from SSE",
				zap.String("client_ip", c.RealIP()))
			return nil
		case <-ticker.C:
			// ดึงข้อมูลเซ็นเซอร์ล่าสุด
			data, err := h.sensorService.GetAllSensors()
			if err != nil {
				h.logger.Error("Failed to get sensor data for SSE update",
					zap.Error(err))
				continue
			}
			// ใช้ timestamp ปัจจุบันเป็น ID
			currentTimestamp = time.Now().Unix()

			// เพิ่ม hostname เข้าไปใน response
			dataWithHostname := fmt.Sprintf(`{"server_id":"%s","data":%s}`, hostname, data)

			// ส่งข้อมูลอัพเดทไปยัง client พร้อม ID
			fmt.Fprintf(c.Response(), "id: %d\nevent: message\ndata: %s\n\n", currentTimestamp, dataWithHostname)
			c.Response().Flush()
		case <-pingTicker.C:
			// ใช้ timestamp ปัจจุบันเป็น ID
			currentTimestamp = time.Now().Unix()
			// ส่ง ping เพื่อให้การเชื่อมต่อยังคงอยู่ พร้อม ID และ hostname
			pingData := fmt.Sprintf(`{"ping": true, "server_id": "%s"}`, hostname)
			fmt.Fprintf(c.Response(), "id: %d\nevent: ping\ndata: %s\n\n", currentTimestamp, pingData)
			c.Response().Flush()
		}
	}
}

// GetSensorData คืนค่าข้อมูล sensor ทั้งหมด
func (h *SensorHandler) GetSensorData(c echo.Context) error {
	data, err := h.sensorService.GetAllSensors()
	if err != nil {
		h.logger.Error("Failed to get sensor data",
			zap.Error(err))
		return apierror.HandleAPIError(c, apierror.Wrap(apierror.ErrDataNotFound, "failed to get sensor data"))
	}

	return c.JSONBlob(http.StatusOK, []byte(data))
}

// GetSensorByID คืนค่าข้อมูล sensor ตาม ID
func (h *SensorHandler) GetSensorByID(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return apierror.HandleAPIError(c, apierror.ErrInvalidRequest)
	}

	sensorJSON, err := h.sensorService.GetSensorByID(id)
	if err != nil {
		h.logger.Error("Failed to get sensor by ID",
			zap.String("id", id),
			zap.Error(err))
		return apierror.HandleAPIError(c, apierror.Wrap(apierror.ErrDataNotFound,
			fmt.Sprintf("sensor with ID %s not found", id)))
	}

	return c.JSONBlob(http.StatusOK, []byte(sensorJSON))
}
