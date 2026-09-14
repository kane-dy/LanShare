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
	// 1. 通过 UDP 建立伪连接，让 OS 底层路由选择最佳出向网卡 IP
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		return localAddr.IP.String(), nil
	}

	// 2. 降级方案：若无外网连接（完全离线环境），遍历物理网卡
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		// 过滤已关闭或回环网卡
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
	// 1. 确保目标保存目录存在（不存在则自动创建）
	if err := os.MkdirAll(a.saveDirectory, 0755); err != nil {
		return err
	}

	// 2. 将路径转换为适用于当前 OS 的规范格式（特别是 Windows 下的 \ 分隔符）
	cleanPath := filepath.Clean(a.saveDirectory)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		// Windows 使用 explorer 打开指定绝对路径
		cmd = exec.Command("explorer", cleanPath)
	case "darwin":
		// macOS 使用 open 打开目录
		cmd = exec.Command("open", cleanPath)
	default: // linux
		// Linux 使用 xdg-open 打开目录
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
	// 用户如果点了取消，filePaths 会是空的 slice []string{}
	return filePaths, nil
}

//// SelectFile 唤起系统原生文件选择弹窗，返回选中的文件绝对路径
//func (a *App) SelectFile() (string, error) {
//	filePath, err := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
//		Title: "选择要发送的文件",
//		Filters: []wailsRuntime.FileFilter{
//			{
//				DisplayName: "所有文件 (*.*)",
//				Pattern:     "*.*",
//			},
//		},
//	})
//	if err != nil {
//		return "", err
//	}
//	return filePath, nil
//}
