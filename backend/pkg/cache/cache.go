package cache

import (
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/Napat/go-sse-sensor-dashboard-demo/backend/pkg/logger"
)

// CacheItem เป็นโครงสร้างที่เก็บข้อมูลและเวลาหมดอายุของ cache
type CacheItem struct {
	value      []byte
	expiration time.Time
}

// IsExpired ตรวจสอบว่า cache item หมดอายุหรือยัง
func (item *CacheItem) IsExpired() bool {
	return time.Now().After(item.expiration)
}

// Cache เป็นโครงสร้างที่ใช้จัดการข้อมูล cache
type Cache struct {
	items map[string]CacheItem
	mu    sync.RWMutex
}

// NewCache สร้าง instance ใหม่ของ Cache
func NewCache() *Cache {
	cache := &Cache{
		items: make(map[string]CacheItem),
	}

	// เริ่ม goroutine สำหรับการล้าง cache ที่หมดอายุ
	go cache.startCleaner()

	return cache
}

// Set บันทึกข้อมูลลงใน cache พร้อมกำหนดเวลาหมดอายุ
func (c *Cache) Set(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = CacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
}

// Get ดึงข้อมูลจาก cache และตรวจสอบอายุ
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	// ตรวจสอบว่า cache หมดอายุหรือยัง
	if item.IsExpired() {
		// ไม่ต้องลบออกตรงนี้ เพราะจะถูกล้างโดย cleaner goroutine
		return nil, false
	}

	return item.value, true
}

// Delete ลบข้อมูลออกจาก cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear ล้างข้อมูลทั้งหมดใน cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]CacheItem)
}

// startCleaner รันในรูปแบบของ goroutine เพื่อล้าง cache ที่หมดอายุเป็นระยะๆ
func (c *Cache) startCleaner() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.cleanExpired()
	}
}

// cleanExpired ล้าง cache ที่หมดอายุออกจากระบบ
func (c *Cache) cleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, item := range c.items {
		if item.IsExpired() {
			delete(c.items, key)
			logger.Debug("Cache entry expired and removed", zap.String("key", key))
		}
	}
}
