package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx           context.Context
	server        *http.Server
	serverLock    sync.Mutex
	saveDirectory string
	configPath    string
	currentMode   string

	sharedFiles []SharedItem
	configLock  sync.RWMutex
}

func NewApp() *App {
	cfgPath := getConfigFilePath()
	savePath := "./LanShareFile"

	userDir, err := os.UserHomeDir()
	if err == nil {
		savePath = filepath.Join(userDir, "Downloads", "LanShareFile")
	}

	app := &App{
		saveDirectory: savePath,
		configPath:    cfgPath,
		sharedFiles:   make([]SharedItem, 0),
	}

	app.loadConfig()
	return app
}

// 🌟 1. 修改 startUp：去除异步协程，确保在 Wails 启动时同步拉起服务
func (a *App) startUp(ctx context.Context) {
	a.ctx = ctx
	_ = os.MkdirAll(a.saveDirectory, 0755)

	// 同步启动服务，打印启动日志以便排查端口或网络问题
	serverUrl, err := a.StartServer(8080)
	if err != nil {
		fmt.Printf("【错误】传输服务自动启动失败: %v\n", err)
	} else {
		fmt.Printf("【成功】传输服务已成功常驻启动: %s\n", serverUrl)
	}
}

// 🌟 2. 补充/修改 GetReceiveUrl：增加防御性兜底，避免前端拿到的 URL 无法访问
func (a *App) GetReceiveUrl() (string, error) {
	a.serverLock.Lock()
	// 防御性保障：如果服务出于某种原因未启动，在此处强制补启动一次
	if a.server == nil {
		a.serverLock.Unlock()
		_, _ = a.StartServer(8080)
	} else {
		a.serverLock.Unlock()
	}

	localIP, err := a.GetLocalIP()
	if err != nil {
		return "", err
	}
	// 明确返回完整的 /upload 上传路径
	return fmt.Sprintf("http://%s:8080/upload", localIP), nil
}

func (a *App) StopServer() error {
	a.serverLock.Lock()
	defer a.serverLock.Unlock()

	if a.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		err := a.server.Shutdown(ctx)
		a.server = nil
		a.currentMode = ""
		return err
	}
	return nil
}

func (a *App) GetSaveDirectory() string {
	return a.saveDirectory
}

func (a *App) SelectSaveDirectory() (string, error) {
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择文件保存目录",
	})

	if err != nil {
		return "", err
	}

	if selection != "" {
		a.saveDirectory = selection
		_ = os.MkdirAll(a.saveDirectory, os.ModePerm)
		a.saveConfig()
		return a.saveDirectory, nil
	}

	return a.saveDirectory, nil
}

func (a *App) SetSaveDirectory(newPath string) error {
	if newPath == "" {
		return fmt.Errorf("路径不能为空")
	}

	err := os.MkdirAll(newPath, os.ModePerm)
	if err != nil {
		return err
	}

	a.saveDirectory = newPath
	a.saveConfig()
	return nil
}
