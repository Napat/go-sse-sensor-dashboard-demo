package service

import (
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/Napat/go-sse-sensor-dashboard-demo/backend/internal/repository"
	"github.com/Napat/go-sse-sensor-dashboard-demo/backend/pkg/cache"
)

const (
	// CacheTTL กำหนดเวลาหมดอายุของ cache
	CacheTTL = 30 * time.Second

	// SensorDataCacheKey คีย์สำหรับ cache ข้อมูลเซนเซอร์ทั้งหมด
	SensorDataCacheKey = "all_sensors"
)

// ISensorService คือ interface สำหรับการเข้าถึงบริการ sensor
type ISensorService interface {
	// GetAllSensors คืนค่าข้อมูล sensor ทั้งหมดในรูปแบบ JSON
	GetAllSensors() (string, error)

	// GetSensorByID คืนค่าข้อมูล sensor ตาม ID ในรูปแบบ JSON
	GetSensorByID(id string) (string, error)
}

// SensorService เป็น implementation ของ ISensorService ที่ใช้ cache
type SensorService struct {
	repository repository.ISensorRepository
	cache      *cache.Cache
	logger     *zap.Logger
}

// NewSensorService สร้าง service ใหม่สำหรับ sensor ที่ใช้ cache
func NewSensorService(repo repository.ISensorRepository, logger *zap.Logger) *SensorService {
	return &SensorService{
		repository: repo,
		cache:      cache.NewCache(),
		logger:     logger,
	}
}

// GetAllSensors คืนค่าข้อมูล sensor ทั้งหมดในรูปแบบ JSON
func (s *SensorService) GetAllSensors() (string, error) {
	// ลองดึงข้อมูลจาก cache ก่อน
	if cachedData, found := s.cache.Get(SensorDataCacheKey); found {
		s.logger.Debug("Cache hit for all sensors data")
		return string(cachedData), nil
	}

	// ถ้าไม่พบใน cache ดึงข้อมูลจริง
	sensors, err := s.repository.GetAllSensors()
	if err != nil {
		s.logger.Error("Failed to get all sensors", zap.Error(err))
		return "", err
	}

	// แปลงเป็น JSON string
	jsonData, err := repository.SerializeSensors(sensors)
	if err != nil {
		s.logger.Error("Failed to serialize sensors", zap.Error(err))
		return "", err
	}

	// เก็บลง cache
	s.cache.Set(SensorDataCacheKey, []byte(jsonData), CacheTTL)
	s.logger.Debug("Cached all sensors data", zap.Duration("ttl", CacheTTL))

	return jsonData, nil
}

// GetSensorByID คืนค่าข้อมูล sensor ตาม ID ในรูปแบบ JSON
func (s *SensorService) GetSensorByID(id string) (string, error) {
	// สร้าง cache key สำหรับ sensor ID นี้
	cacheKey := "sensor_" + id

	// ลองดึงข้อมูลจาก cache ก่อน
	if cachedData, found := s.cache.Get(cacheKey); found {
		s.logger.Debug("Cache hit for sensor", zap.String("id", id))
		return string(cachedData), nil
	}

	// ถ้าไม่พบใน cache ดึงข้อมูลจริง
	sensor, err := s.repository.GetSensorByID(id)
	if err != nil {
		s.logger.Error("Failed to get sensor by ID", zap.String("id", id), zap.Error(err))
		return "", err
	}

	// แปลงเป็น JSON string
	jsonData, err := repository.SerializeSensor(sensor)
	if err != nil {
		s.logger.Error("Failed to serialize sensor", zap.String("id", id), zap.Error(err))
		return "", err
	}

	// เก็บลง cache
	s.cache.Set(cacheKey, []byte(jsonData), CacheTTL)
	s.logger.Debug("Cached sensor data", zap.String("id", id), zap.Duration("ttl", CacheTTL))

	return jsonData, nil
}

// SensorServiceInstance กำหนดตัวแปรสำหรับ singleton pattern
var (
	sensorServiceInstance ISensorService
	sensorServiceOnce     sync.Once
)

// GetSensorService คืนค่า instance ของ ISensorService แบบ singleton
func GetSensorService(logger *zap.Logger) ISensorService {
	sensorServiceOnce.Do(func() {
		repo := repository.NewSensorRepository()
		sensorServiceInstance = NewSensorService(repo, logger)
	})
	return sensorServiceInstance
}
