package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SharedItem 代表单条分享记录
type SharedItem struct {
	ID        string    `json:"id"`
	FileName  string    `json:"file_name"`
	FilePath  string    `json:"file_path"`
	FileSize  int64     `json:"file_size"`
	CreatedAt time.Time `json:"created_at"`
}

// Config 配置文件的持久化结构体
type Config struct {
	SaveDirectory string       `json:"save_directory"`
	SharedFiles   []SharedItem `json:"shared_files"`
}

// 获取配置文件的路径
func getConfigFilePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	appConfigDir := filepath.Join(configDir, "LanShare")
	_ = os.MkdirAll(appConfigDir, 0755)
	return filepath.Join(appConfigDir, "config.json")
}

// loadConfig 读取本地配置文件
func (a *App) loadConfig() {
	a.configLock.Lock()
	defer a.configLock.Unlock()

	data, err := os.ReadFile(a.configPath)
	if err != nil {
		return
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err == nil {
		if cfg.SaveDirectory != "" {
			a.saveDirectory = cfg.SaveDirectory
		}
		if cfg.SharedFiles != nil {
			a.sharedFiles = cfg.SharedFiles
		}
	}
}

// saveConfigLocked 未加锁的保存（由外部函数控制锁）
func (a *App) saveConfigLocked() {
	cfg := Config{
		SaveDirectory: a.saveDirectory,
		SharedFiles:   a.sharedFiles,
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err == nil {
		_ = os.WriteFile(a.configPath, data, 0644)
	}
}

// saveConfig 带锁保存
func (a *App) saveConfig() {
	a.configLock.Lock()
	defer a.configLock.Unlock()
	a.saveConfigLocked()
}

// AddSharedFile 添加分享文件
func (a *App) AddSharedFile(item SharedItem) error {
	// 🌟 1. 校验文件路径
	if item.FilePath == "" {
		return fmt.Errorf("文件路径不能为空")
	}

	// 🌟 2. 在后台通过 os.Stat 读取文件信息并计算大小
	fi, err := os.Stat(item.FilePath)
	if err != nil {
		return fmt.Errorf("无法获取文件信息或文件不存在: %v", err)
	}

	if fi.IsDir() {
		return fmt.Errorf("暂不支持添加文件夹，请选择单文件")
	}

	// 🌟 3. 自动计算文件大小（字节数）与文件名
	item.FileSize = fi.Size()
	if item.FileName == "" {
		item.FileName = fi.Name()
	}

	a.configLock.Lock()
	defer a.configLock.Unlock()

	if item.ID == "" {
		item.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}

	exists := false
	for i, existing := range a.sharedFiles {
		if existing.FilePath == item.FilePath {
			a.sharedFiles[i] = item
			exists = true
			break
		}
	}
	if !exists {
		a.sharedFiles = append(a.sharedFiles, item)
	}

	a.saveConfigLocked()
	return nil
}

// GetSharedFiles 获取分享列表
func (a *App) GetSharedFiles() []SharedItem {
	a.configLock.Lock()
	defer a.configLock.Unlock()

	result := make([]SharedItem, len(a.sharedFiles))
	copy(result, a.sharedFiles)
	return result
}

// DeleteSharedFile 删除分享记录
func (a *App) DeleteSharedFile(id string) error {
	a.configLock.Lock()
	defer a.configLock.Unlock()

	newList := make([]SharedItem, 0)
	for _, item := range a.sharedFiles {
		if item.ID != id {
			newList = append(newList, item)
		}
	}
	a.sharedFiles = newList
	a.saveConfigLocked()
	return nil
}
