// Package logger 提供 CraftFire 统一的日志系统。
// 支持日志级别控制和格式化输出。
package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// 日志级别常量
const (
	LevelDebug = iota
	LevelInfo
	LevelWarn
	LevelError
)

var (
	currentLevel = LevelInfo
	logger       = log.New(os.Stdout, "", 0)
)

// Init 初始化日志系统，设置日志级别。
// 参数：
//   - level: 日志级别字符串 ("DEBUG", "INFO", "WARN", "ERROR")
func Init(level string) {
	switch strings.ToUpper(level) {
	case "DEBUG":
		currentLevel = LevelDebug
	case "INFO":
		currentLevel = LevelInfo
	case "WARN":
		currentLevel = LevelWarn
	case "ERROR":
		currentLevel = LevelError
	default:
		currentLevel = LevelInfo
	}
}

// formatMsg 格式化日志消息，添加时间戳和级别标签。
func formatMsg(level string, format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf(format, args...)
	return fmt.Sprintf("[%s] [%s] %s", timestamp, level, msg)
}

// Debug 输出调试级别日志。
func Debug(format string, args ...interface{}) {
	if currentLevel <= LevelDebug {
		logger.Println(formatMsg("DEBUG", format, args...))
	}
}

// Info 输出信息级别日志。
func Info(format string, args ...interface{}) {
	if currentLevel <= LevelInfo {
		logger.Println(formatMsg("INFO", format, args...))
	}
}

// Warn 输出警告级别日志。
func Warn(format string, args ...interface{}) {
	if currentLevel <= LevelWarn {
		logger.Println(formatMsg("WARN", format, args...))
	}
}

// Error 输出错误级别日志。
func Error(format string, args ...interface{}) {
	if currentLevel <= LevelError {
		logger.Println(formatMsg("ERROR", format, args...))
	}
}
