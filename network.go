package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// GetLocalIP 优先通过 UDP 获取主网卡 IPv4，获取失败时降级遍历网卡
func (a *App) GetLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP.String(), nil
	}

	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ip4 := ipnet.IP.To4(); ip4 != nil {
					return ip4.String(), nil
				}
			}
		}
	}

	return "", fmt.Errorf("未找到有效的局域网 IP 地址")
}

// OpenSaveDirectory 调用系统默认文件管理器打开接收文件夹
func (a *App) OpenSaveDirectory() error {
	if err := os.MkdirAll(a.saveDirectory, 0755); err != nil {
		return err
	}

	cleanPath := filepath.Clean(a.saveDirectory)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", cleanPath)
	case "darwin":
		cmd = exec.Command("open", cleanPath)
	default: // linux
		cmd = exec.Command("xdg-open", cleanPath)
	}

	return cmd.Start()
}

// SelectFile 唤起系统原生文件选择弹窗，返回选中的所有文件绝对路径
func (a *App) SelectFile() ([]string, error) {
	filePaths, err := wailsRuntime.OpenMultipleFilesDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "选择要发送的文件（可多选）",
		Filters: []wailsRuntime.FileFilter{
			{
				DisplayName: "所有文件 (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return filePaths, nil
}

// SelectDirectory 唤起系统原生文件夹选择弹窗，返回选中的文件夹绝对路径
func (a *App) SelectDirectory() (string, error) {
	dirPath, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "选择要发送的文件夹",
	})
	if err != nil {
		return "", err
	}
	return dirPath, nil
}
