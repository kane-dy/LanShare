package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// 定义配置文件的持久化结构体
type Config struct {
	SaveDirectory string `json:"save_directory"`
}

// 定义 LanShare 应用的核心数据模型与服务状态
type App struct {
	ctx           context.Context
	server        *http.Server // HTTP 服务实例
	serverLock    sync.Mutex   // 保证启动/停止操作的并发安全
	saveDirectory string       // 本地接收文件的保存目录
	configPath    string       // 本地配置文件绝对路径
	currentMode   string       // "send" 或 "receive"
}

// 获取持久化配置文件的保存位置（存储在系统的 AppData/Config 目录下）
func getConfigFilePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}
	appConfigDir := filepath.Join(configDir, "LanShare")
	_ = os.MkdirAll(appConfigDir, 0755)
	return filepath.Join(appConfigDir, "config.json")
}

func NewApp() *App {
	// 1. 获取配置文件保存路径
	cfgPath := getConfigFilePath()

	// 2. 默认降级目录（如果无法获取用户家目录，保存在当前程序运行目录下的 LanShareFile）
	savePath := "./LanShareFile"

	// 获取当前系统登录的用户家目录
	userDir, err := os.UserHomeDir()
	if err == nil {
		// 拼接系统 Downloads 目录
		savePath = filepath.Join(userDir, "Downloads", "LanShareFile")
	}

	app := &App{
		saveDirectory: savePath,
		configPath:    cfgPath,
	}

	// 3. 读取本地历史配置（如果有修改记录，自动覆盖默认保存目录）
	app.loadConfig()

	return app
}

// 读取本地配置文件
func (a *App) loadConfig() {
	data, err := os.ReadFile(a.configPath)
	if err != nil {
		return // 文件不存在或读取失败时，直接保留默认 saveDirectory 路径
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err == nil && cfg.SaveDirectory != "" {
		a.saveDirectory = cfg.SaveDirectory
	}
}

// 保存当前配置到本地 JSON 文件
func (a *App) saveConfig() {
	cfg := Config{
		SaveDirectory: a.saveDirectory,
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err == nil {
		_ = os.WriteFile(a.configPath, data, 0644)
	}
}

// Wails 框架在应用完成初始化后自动调用的钩子函数
func (a *App) startUp(ctx context.Context) {
	// 传递进来的 context, 供全局使用
	a.ctx = ctx
	// 应用启动时自动创建目标保存文件夹
	_ = os.MkdirAll(a.saveDirectory, 0755)
}

// 用于关闭 HTTP 服务
func (a *App) StopServer() error {
	// 加锁
	a.serverLock.Lock()
	// 自动释放锁
	defer a.serverLock.Unlock()
	// 检查是否有启动的 HTTP 服务
	if a.server != nil {
		// 创建一个带有 2 秒超时的上下文，确保服务关闭过程不会无限期卡死
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		// 延迟调用 cancel 函数，释放与超时上下文相关的内存资源
		defer cancel()
		// 调用 Shutdown 方法优雅地关闭 HTTP 服务，等待当前正在传输的数据完成后再断开
		err := a.server.Shutdown(ctx)
		// 将服务实例指针置空，标志当前服务已彻底停止
		a.server = nil
		// 重置当前运行模式标记为空字符串
		a.currentMode = ""
		// 返回 Shutdown 关闭服务时产生的错误状态（若正常关闭则为 nil）
		return err
	}
	// 如果本来就没有运行中的服务，直接返回 nil
	return nil
}

// 1. 获取当前设置的保存目录
func (a *App) GetSaveDirectory() string {
	return a.saveDirectory
}

// 2. 打开系统文件选择器，让用户重新选择保存文件夹
func (a *App) SelectSaveDirectory() (string, error) {
	// 调用 Wails 原生文件夹选择对话框
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择文件保存目录",
	})

	if err != nil {
		return "", err
	}

	// 如果用户选择了路径（没有取消选择）
	if selection != "" {
		a.saveDirectory = selection
		// 确保选中的目录存在
		_ = os.MkdirAll(a.saveDirectory, os.ModePerm)

		// 👈 更改路径后保存至本地配置文件
		a.saveConfig()

		return a.saveDirectory, nil
	}

	return a.saveDirectory, nil
}

// 3. 手动设置/修改保存目录路径
func (a *App) SetSaveDirectory(newPath string) error {
	if newPath == "" {
		return fmt.Errorf("路径不能为空")
	}

	// 创建目录（如果不存在）
	err := os.MkdirAll(newPath, os.ModePerm)
	if err != nil {
		return err
	}

	a.saveDirectory = newPath

	// 👈 更改路径后保存至本地配置文件
	a.saveConfig()

	return nil
}
