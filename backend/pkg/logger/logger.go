package logger

import (
	"os"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log  *zap.Logger
	once sync.Once
)

// Init ตั้งค่า logger ตามสภาพแวดล้อม
func Init(env string) *zap.Logger {
	return initLogger(env, "")
}

// InitTempLogger สร้าง logger ชั่วคราวสำหรับใช้ระหว่างการโหลด config
func InitTempLogger(env string) *zap.Logger {
	// ไม่ใช้ once.Do เพื่อให้สามารถสร้าง logger ใหม่ได้
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var config zap.Config
	if env == "dev" {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig = encoderConfig
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig = encoderConfig
	}

	logger, err := config.Build()
	if err != nil {
		return zap.NewExample()
	}

	return logger
}

// InitWithLevel สร้าง logger ด้วย level ที่กำหนด
func InitWithLevel(env string, logLevel string) *zap.Logger {
	// รีเซ็ต once เพื่อให้สามารถสร้าง logger ใหม่ได้
	once = sync.Once{}
	return initLogger(env, logLevel)
}

// initLogger ฟังก์ชันภายในสำหรับสร้าง logger
func initLogger(env string, logLevelOverride string) *zap.Logger {
	once.Do(func() {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		var config zap.Config
		if env == "dev" {
			config = zap.NewDevelopmentConfig()
			config.EncoderConfig = encoderConfig
			config.OutputPaths = []string{"stdout"}
			config.ErrorOutputPaths = []string{"stderr"}
			config.DisableCaller = false
			config.DisableStacktrace = false
		} else {
			config = zap.NewProductionConfig()
			config.EncoderConfig = encoderConfig
			config.OutputPaths = []string{"stdout"}
			config.ErrorOutputPaths = []string{"stderr"}
			config.DisableCaller = true
			config.DisableStacktrace = false
		}

		// ใช้ค่า log level ที่ระบุโดยตรง (จาก config)
		if logLevelOverride != "" {
			var level zapcore.Level
			switch strings.ToLower(logLevelOverride) {
			case "debug":
				level = zapcore.DebugLevel
			case "info":
				level = zapcore.InfoLevel
			case "warn":
				level = zapcore.WarnLevel
			case "error":
				level = zapcore.ErrorLevel
			default:
				level = zapcore.InfoLevel
			}
			config.Level = zap.NewAtomicLevelAt(level)
		} else {
			// หรือใช้ค่าจาก environment variable
			logLevelEnv := os.Getenv("APP_LOG_LEVEL")
			if logLevelEnv != "" {
				var level zapcore.Level
				switch strings.ToLower(logLevelEnv) {
				case "debug":
					level = zapcore.DebugLevel
				case "info":
					level = zapcore.InfoLevel
				case "warn":
					level = zapcore.WarnLevel
				case "error":
					level = zapcore.ErrorLevel
				default:
					level = zapcore.InfoLevel
				}
				config.Level = zap.NewAtomicLevelAt(level)
			}
		}

		logger, err := config.Build()
		if err != nil {
			Log = zap.NewExample()
			Log.Error("Failed to initialize zap logger", zap.Error(err))
			return
		}

		Log = logger
		Log.Info("Logger initialized", zap.String("environment", env), zap.String("log_level", config.Level.String()))
	})

	return Log
}

// GetLogger คืนค่า logger ที่มีอยู่หรือสร้างใหม่ถ้ายังไม่มี
func GetLogger() *zap.Logger {
	if Log == nil {
		env := os.Getenv("APP_ENV")
		if env == "" {
			env = "dev"
		}
		return Init(env)
	}
	return Log
}

// Sync เรียก sync ของ logger เพื่อให้แน่ใจว่า log ทั้งหมดจะถูกเขียน
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}

// Info บันทึก log ระดับ Info
func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

// Debug บันทึก log ระดับ Debug
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

// Warn บันทึก log ระดับ Warn
func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

// Error บันทึก log ระดับ Error
func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

// Fatal บันทึก log ระดับ Fatal และจบการทำงาน
func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// With สร้าง logger ใหม่พร้อม fields เพิ่มเติม
func With(fields ...zap.Field) *zap.Logger {
	return GetLogger().With(fields...)
}

// LogRequestInfo บันทึก log การเรียกใช้ API
func LogRequestInfo(method, path, ip, userAgent string, statusCode int, latency float64, fields ...zap.Field) {
	allFields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.String("ip", ip),
		zap.Int("status", statusCode),
		zap.Float64("latency_ms", latency),
		zap.String("user_agent", userAgent),
	}

	if len(fields) > 0 {
		allFields = append(allFields, fields...)
	}

	if statusCode >= 500 {
		Error("Request failed", allFields...)
	} else if statusCode >= 400 {
		Warn("Request warning", allFields...)
	} else {
		Info("Request completed", allFields...)
	}
}
