package model

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

// ValidateFile 验证 3D 模型文件的有效性。
// 检查文件格式和大小。
//
// 参数：
//   - filePath: 文件路径
func ValidateFile(filePath string) error {
	// 验证格式
	if err := ValidateFormat(filePath); err != nil {
		return err
	}

	// 验证文件大小
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("无法读取文件: %w", err)
	}

	if fileInfo.Size() > maxFileSize {
		return fmt.Errorf("文件大小超限: %d 字节 (最大 %d 字节)", fileInfo.Size(), maxFileSize)
	}

	if fileInfo.Size() == 0 {
		return fmt.Errorf("文件为空")
	}

	return nil
}

// CalculateMD5 计算文件的 MD5 哈希值。
//
// 参数：
//   - filePath: 文件路径
//
// 返回值：MD5 哈希字符串（十六进制）
func CalculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("计算哈希失败: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
