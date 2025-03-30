package repository

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/Napat/go-sse-sensor-dashboard-demo/backend/internal/model"
)

// ISensorRepository คือ interface สำหรับการเข้าถึงข้อมูล sensor
type ISensorRepository interface {
	// GetAllSensors คืนค่าข้อมูล sensor ทั้งหมด
	GetAllSensors() ([]*model.SensorModel, error)

	// GetSensorByID คืนค่าข้อมูล sensor ตาม ID
	GetSensorByID(id string) (*model.SensorModel, error)

	// UpdateRandomSensorData อัปเดตข้อมูลเซนเซอร์แบบสุ่ม
	UpdateRandomSensorData()
}

// SensorRepository เป็น implementation ของ ISensorRepository ที่ใช้ข้อมูลจำลอง
type SensorRepository struct {
	sensors map[string]*model.SensorModel
	mutex   sync.RWMutex
}

// NewSensorRepository สร้าง repository ใหม่สำหรับ sensor
func NewSensorRepository() *SensorRepository {
	repo := &SensorRepository{
		sensors: make(map[string]*model.SensorModel),
	}

	// สร้างข้อมูลจำลอง
	repo.initMockSensors()

	// เริ่มต้น goroutine สำหรับการจำลองข้อมูลเซนเซอร์
	go repo.mockSensorDataLoop()

	return repo
}

// initMockSensors สร้างข้อมูลเซนเซอร์จำลอง
func (r *SensorRepository) initMockSensors() {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// สร้างเซนเซอร์จำลอง
	r.sensors["temp-001"] = &model.SensorModel{
		ID:          "temp-001",
		Name:        "Temperature Sensor 1",
		Type:        "temperature",
		Temperature: 25.0,
		Humidity:    0,
		Timestamp:   time.Now(),
		Status:      "active",
	}

	r.sensors["temp-002"] = &model.SensorModel{
		ID:          "temp-002",
		Name:        "Temperature Sensor 2",
		Type:        "temperature",
		Temperature: 22.5,
		Humidity:    0,
		Timestamp:   time.Now(),
		Status:      "active",
	}

	r.sensors["humid-001"] = &model.SensorModel{
		ID:          "humid-001",
		Name:        "Humidity Sensor 1",
		Type:        "humidity",
		Temperature: 0,
		Humidity:    45.0,
		Timestamp:   time.Now(),
		Status:      "active",
	}

	r.sensors["combined-001"] = &model.SensorModel{
		ID:          "combined-001",
		Name:        "Combined Sensor 1",
		Type:        "combined",
		Temperature: 24.0,
		Humidity:    40.0,
		Timestamp:   time.Now(),
		Status:      "active",
	}
}

// GetAllSensors คืนค่าข้อมูล sensor ทั้งหมด
func (r *SensorRepository) GetAllSensors() ([]*model.SensorModel, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// สร้าง slice จากข้อมูล map
	sensors := make([]*model.SensorModel, 0, len(r.sensors))
	for _, sensor := range r.sensors {
		sensors = append(sensors, sensor)
	}

	return sensors, nil
}

// GetSensorByID คืนค่าข้อมูล sensor ตาม ID
func (r *SensorRepository) GetSensorByID(id string) (*model.SensorModel, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// ค้นหาเซนเซอร์ตาม ID
	sensor, ok := r.sensors[id]
	if !ok {
		return nil, fmt.Errorf("sensor with ID %s not found", id)
	}

	return sensor, nil
}

// UpdateRandomSensorData อัปเดตข้อมูลเซนเซอร์แบบสุ่ม
func (r *SensorRepository) UpdateRandomSensorData() {
	// สุ่มค่าอุณหภูมิ 20-30°C และความชื้น 30-50%
	temp := 20 + rand.Float64()*10
	humidity := 30 + rand.Float64()*20

	r.mutex.Lock()
	defer r.mutex.Unlock()

	// อัปเดตค่าให้กับเซนเซอร์ทั้งหมด
	now := time.Now()
	for _, sensor := range r.sensors {
		sensor.Timestamp = now

		switch sensor.Type {
		case "temperature":
			// ให้มีความแตกต่างเล็กน้อยระหว่างเซนเซอร์
			sensor.Temperature = temp + (rand.Float64()-0.5)*2
		case "humidity":
			// ให้มีความแตกต่างเล็กน้อยระหว่างเซนเซอร์
			sensor.Humidity = humidity + (rand.Float64()-0.5)*5
		case "combined":
			// สำหรับเซนเซอร์แบบรวม อัปเดตทั้งอุณหภูมิและความชื้น
			sensor.Temperature = temp + (rand.Float64()-0.5)*2
			sensor.Humidity = humidity + (rand.Float64()-0.5)*5
		}
	}
}

// mockSensorDataLoop ใช้สำหรับสุ่มค่าเซนเซอร์เป็นระยะ
func (r *SensorRepository) mockSensorDataLoop() {
	for {
		// อัปเดตข้อมูลแบบสุ่ม
		r.UpdateRandomSensorData()

		// รอ 2 วินาทีก่อนอัปเดตค่าถัดไป
		time.Sleep(2 * time.Second)
	}
}

// SerializeSensor แปลงข้อมูล SensorModel เป็น JSON string
func SerializeSensor(sensor *model.SensorModel) (string, error) {
	data, err := json.Marshal(sensor)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SerializeSensors แปลงข้อมูล slice ของ SensorModel เป็น JSON string
func SerializeSensors(sensors []*model.SensorModel) (string, error) {
	data, err := json.Marshal(sensors)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
