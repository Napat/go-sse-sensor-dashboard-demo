package model

import (
	"time"
)

// SensorModel คือข้อมูลของเซนเซอร์ที่จะส่งกลับไปให้ client
type SensorModel struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Temperature float64   `json:"temperature"`
	Humidity    float64   `json:"humidity"`
	Timestamp   time.Time `json:"timestamp"`
	Status      string    `json:"status"`
}
