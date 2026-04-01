package model

import (
	"io"
	"os"

	applogger "CraftFire/backend/internal/logger"
)

// CopyFile 将源文件复制到目标路径。
//
// 参数：
//   - src: 源文件路径
//   - dst: 目标文件路径
func CopyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	applogger.Debug("文件已复制: %s -> %s", src, dst)
	return nil
}
