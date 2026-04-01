package model

import (
	"fmt"
	"path/filepath"
	"strings"

	applogger "CraftFire/backend/internal/logger"
)

// 支持的 3D 模型格式
var supportedFormats = map[string]bool{
	".gltf": true,
	".glb":  true,
	".fbx":  true,
	".obj":  true,
	".dae":  true,
}

// maxFileSize 单个模型文件最大大小（10MB）
const maxFileSize int64 = 10 * 1024 * 1024

// Importer 3D 模型导入处理器。
type Importer struct{}

// NewImporter 创建一个新的模型导入器。
func NewImporter() *Importer {
	return &Importer{}
}

// ValidateFormat 验证文件格式是否支持。
func ValidateFormat(filePath string) error {
	ext := strings.ToLower(filepath.Ext(filePath))
	if !supportedFormats[ext] {
		return fmt.Errorf("不支持的文件格式: %s (支持: GLTF, GLB, FBX, OBJ, DAE)", ext)
	}
	applogger.Debug("文件格式验证通过: %s", ext)
	return nil
}

// GetSupportedFormats 获取支持的文件格式列表。
func GetSupportedFormats() []string {
	formats := make([]string, 0, len(supportedFormats))
	for ext := range supportedFormats {
		formats = append(formats, ext)
	}
	return formats
}
