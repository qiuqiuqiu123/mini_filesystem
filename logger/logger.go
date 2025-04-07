package logger

import (
	"log"
	"os"
	"sync"
)

type LogLevel int

const callDepth = 4

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Logger struct {
	level       LogLevel    // 当前日志级别
	debugLogger *log.Logger // 不同级别的 Logger 实例
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
}

var (
	instance *Logger
	once     sync.Once
)

func Init(level LogLevel) {
	once.Do(func() {
		// 创建不同级别的 Logger（可自定义输出目标）
		instance = &Logger{
			level:       level,
			debugLogger: log.New(&logWriter{calldepth: callDepth, writer: os.Stdout}, "[DEBUG] ", log.LstdFlags),
			infoLogger:  log.New(&logWriter{calldepth: callDepth, writer: os.Stdout}, "[INFO]  ", log.LstdFlags),
			warnLogger:  log.New(&logWriter{calldepth: callDepth, writer: os.Stdout}, "[WARN]  ", log.LstdFlags),
			errorLogger: log.New(&logWriter{calldepth: callDepth, writer: os.Stdout}, "[ERROR] ", log.LstdFlags),
		}
	})
}

// GetLogger 获取全局单例实例
func GetLogger() *Logger {
	if instance == nil {
		panic("Logger not initialized. Call logger.Init() first.")
	}
	return instance
}

func (l *Logger) Debug(v ...interface{}) {
	if l.level <= LevelDebug {
		l.debugLogger.Println(v...)
	}
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.level <= LevelDebug {
		l.debugLogger.Printf(format, v...)
	}
}

func (l *Logger) Info(v ...interface{}) {
	if l.level <= LevelInfo {
		l.infoLogger.Println(v...)
	}
}

func (l *Logger) Infof(format string, v ...interface{}) {
	if l.level <= LevelInfo {
		l.infoLogger.Printf(format, v...)
	}
}

func (l *Logger) Warn(v ...interface{}) {
	if l.level <= LevelWarn {
		l.warnLogger.Println(v...)
	}
}

func (l *Logger) Error(v ...interface{}) {
	if l.level <= LevelError {
		l.errorLogger.Println(v...)
	}
}

// 支持格式化输出（如 Errorf）
func (l *Logger) Errorf(format string, v ...interface{}) {
	if l.level <= LevelError {
		l.errorLogger.Printf(format, v...)
	}
}
