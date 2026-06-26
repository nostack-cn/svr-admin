package upload

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Uploader 文件上传器
type Uploader struct {
	BaseDir string
	BaseURL string
	MaxSize int64
}

// NewUploader 创建上传器
func NewUploader(baseDir, baseURL string, maxSize int64) *Uploader {
	return &Uploader{
		BaseDir: baseDir,
		BaseURL: baseURL,
		MaxSize: maxSize,
	}
}

// FileInfo 上传文件信息
type FileInfo struct {
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	MimeType string `json:"mime_type"`
	URL      string `json:"url"`
}

// Save 保存上传文件
func (u *Uploader) Save(file *multipart.FileHeader, subDir string) (*FileInfo, error) {
	if u.MaxSize > 0 && file.Size > u.MaxSize {
		return nil, fmt.Errorf("文件大小超过限制 (%dMB)", u.MaxSize/1024/1024)
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowed := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
		".svg": true, ".bmp": true,
		".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".ppt": true, ".pptx": true,
		".zip": true, ".rar": true, ".7z": true, ".gz": true,
		".mp4": true, ".mp3": true, ".wav": true,
		".txt": true, ".csv": true, ".json": true, ".xml": true, ".md": true,
	}
	if !allowed[ext] {
		return nil, fmt.Errorf("不支持的文件类型: %s", ext)
	}

	b := make([]byte, 16)
	rand.Read(b)
	newName := hex.EncodeToString(b) + ext

	dateDir := time.Now().Format("2006/01/02")
	targetDir := filepath.Join(u.BaseDir, subDir, dateDir)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	targetPath := filepath.Join(targetDir, newName)
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(targetPath)
	if err != nil {
		return nil, fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	urlPath := fmt.Sprintf("%s/%s/%s/%s", u.BaseURL, subDir, dateDir, newName)
	return &FileInfo{
		FileName: file.Filename,
		FileSize: file.Size,
		MimeType: file.Header.Get("Content-Type"),
		URL:      urlPath,
	}, nil
}
