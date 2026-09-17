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
	configLock  sync.Mutex
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

func (a *App) startUp(ctx context.Context) {
	a.ctx = ctx
	_ = os.MkdirAll(a.saveDirectory, 0755)
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
