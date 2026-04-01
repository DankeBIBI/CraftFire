// Package utils 提供通用工具函数，包括 UUID 生成等。
package utils

import (
	"crypto/rand"
	"fmt"
)

// GenerateUUID 生成一个标准的 UUID v4 字符串。
// 返回格式：xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
func GenerateUUID() string {
	uuid := make([]byte, 16)
	_, _ = rand.Read(uuid)

	// 设置版本 4
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	// 设置变体
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

// GenerateShortID 生成指定长度的短 ID（仅包含十六进制字符）。
func GenerateShortID(length int) string {
	bytes := make([]byte, length)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("%x", bytes)[:length]
}
