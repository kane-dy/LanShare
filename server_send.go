package main

import (
	"fmt"
	"net/url"
	"os"
)

// StartSendMode 开启发文件模式（将选择的文件/文件夹加入统一服务并生成下载链接）
func (a *App) StartSendMode(filePaths []string, port int) (string, error) {
	if len(filePaths) == 0 {
		return "", fmt.Errorf("未选择任何文件")
	}

	// 1. 确保统一服务正在运行（不打断接收和分享库服务）
	_, err := a.StartServer(port)
	if err != nil {
		return "", err
	}

	var targetFilePath string

	// 2. 判断是单文件还是多文件/文件夹
	if len(filePaths) == 1 {
		info, err := os.Stat(filePaths[0])
		if err != nil {
			return "", fmt.Errorf("文件不存在: %v", err)
		}

		if !info.IsDir() {
			// 单文件直接使用路径
			targetFilePath = filePaths[0]
		} else {
			// 单个文件夹：打包为 zip（逻辑可根据需求扩展）
			return "", fmt.Errorf("暂不支持直接发送文件夹，请选择单文件或将文件夹加入分享库")
		}
	} else {
		// 多文件发送逻辑：建议提示用户直接加入分享库，或选单文件
		return "", fmt.Errorf("多文件请直接添加至【局域网分享库】中统一管理与下载")
	}

	// 3. 将该文件自动添加进分享库列表（消除变量未使用报错，并实现统一管理）
	err = a.AddSharedFile(SharedItem{
		FilePath: targetFilePath,
	})
	if err != nil {
		return "", fmt.Errorf("添加传输文件失败: %v", err)
	}

	localIP, err := a.GetLocalIP()
	if err != nil {
		return "", err
	}

	// 4. 获取文件信息拼装分享库的下载 URL
	fi, err := os.Stat(targetFilePath)
	if err != nil {
		return "", err
	}

	downloadUrl := fmt.Sprintf("http://%s:%d/download/shared/%s?path=%s",
		localIP,
		port,
		url.PathEscape(fi.Name()),
		url.QueryEscape(targetFilePath),
	)

	return downloadUrl, nil
}

//func zipFileToArchive(zw *zip.Writer, srcPath, zipPath string) error {
//	file, err := os.Open(srcPath)
//	if err != nil {
//		return err
//	}
//	defer file.Close()
//
//	w, err := zw.Create(zipPath)
//	if err != nil {
//		return err
//	}
//
//	_, err = io.Copy(w, file)
//	return err
//}

// StartShareHubMode 启动全局分享库模式（返回分享库列表页面地址，不关闭现有服务）
func (a *App) StartShareHubMode(port int) (string, error) {
	// 🌟 1. 启动或复用统一服务
	baseUrl, err := a.StartServer(port)
	if err != nil {
		return "", err
	}

	// 🌟 2. 返回根路径地址给分享库生成二维码
	return baseUrl, nil
}
