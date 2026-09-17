package main

import (
	"fmt"
)

// StartReceiveMode 开启收文件模式（返回接收上传页面地址，不关闭现有服务）
func (a *App) StartReceiveMode(port int) (string, error) {
	// 🌟 1. 启动或复用统一服务
	_, err := a.StartServer(port)
	if err != nil {
		return "", err
	}

	localIP, err := a.GetLocalIP()
	if err != nil {
		return "", err
	}

	// 🌟 2. 返回专属的 /upload 上传页面 URL 给接收界面生成二维码
	return fmt.Sprintf("http://%s:%d/upload", localIP, port), nil
}
