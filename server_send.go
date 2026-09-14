package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// StartSendMode 开启发文件模式（支持单文件/多文件/文件夹，自动打 ZIP 包）
func (a *App) StartSendMode(filePaths []string, port int) (string, error) {
	_ = a.StopServer()

	a.serverLock.Lock()
	defer a.serverLock.Unlock()

	if len(filePaths) == 0 {
		return "", fmt.Errorf("未选择任何文件")
	}

	// 1. 判断是单文件还是多文件
	var downloadFileName string
	var isSingleFile bool = false
	var singleFilePath string

	if len(filePaths) == 1 {
		info, err := os.Stat(filePaths[0])
		if err != nil {
			return "", fmt.Errorf("文件不存在: %v", err)
		}
		// 如果选中的是单个普通文件（不是文件夹）
		if !info.IsDir() {
			isSingleFile = true
			singleFilePath = filePaths[0]
			downloadFileName = info.Name()
		}
	}

	// 如果是多个文件或包含文件夹，统一打包命名为 "LanShare_传输文件.zip"
	if !isSingleFile {
		downloadFileName = fmt.Sprintf("LanShare_传输文件_%s.zip", time.Now().Format("150405"))
	}

	mux := http.NewServeMux()

	// 2. 构造干净的下载路由（例如 /download/file.zip 或 /download/file.docx）
	ext := filepath.Ext(downloadFileName)
	if ext == "" && !isSingleFile {
		ext = ".zip"
	}
	cleanRoute := fmt.Sprintf("/download/file%s", ext)

	// 3. 注册下载处理路由
	mux.HandleFunc("/download/", func(w http.ResponseWriter, r *http.Request) {
		// 防缓存标头
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		// 强制字节流下载，禁止浏览器在线预览
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// 设 Content-Disposition 标头（正确处理中文文件名）
		encodedFileName := url.QueryEscape(downloadFileName)
		contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", encodedFileName, encodedFileName)
		w.Header().Set("Content-Disposition", contentDisposition)

		// 情况 1：如果是单个文件，直接高效流式读取发送
		if isSingleFile {
			file, err := os.Open(singleFilePath)
			if err != nil {
				http.Error(w, "无法读取文件", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			fi, err := file.Stat()
			if err == nil {
				w.Header().Set("Content-Length", fmt.Sprintf("%d", fi.Size()))
				http.ServeContent(w, r, downloadFileName, fi.ModTime(), file)
			}
			return
		}

		// 情况 2：多文件/文件夹，实时流式 Zip 打包输出
		zw := zip.NewWriter(w)
		defer zw.Close()

		for _, path := range filePaths {
			info, err := os.Stat(path)
			if err != nil {
				continue
			}

			if info.IsDir() {
				// 递归打包文件夹
				baseDir := filepath.Dir(path)
				_ = filepath.Walk(path, func(filePath string, fi os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					relPath, err := filepath.Rel(baseDir, filePath)
					if err != nil {
						return err
					}
					// 统一 Zip 内的分隔符为 "/"
					relPath = filepath.ToSlash(relPath)

					if fi.IsDir() {
						if !strings.HasSuffix(relPath, "/") {
							relPath += "/"
						}
						_, err := zw.CreateHeader(&zip.FileHeader{Name: relPath, Method: zip.Deflate})
						return err
					}

					return zipFileToArchive(zw, filePath, relPath)
				})
			} else {
				// 打包单个文件
				_ = zipFileToArchive(zw, path, info.Name())
			}
		}
	})

	a.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	a.currentMode = "send"

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("发文件服务异常终止: %v\n", err)
		}
	}()

	localIP, err := a.GetLocalIP()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%d%s?t=%d", localIP, port, cleanRoute, time.Now().UnixNano()), nil
}

// 辅助函数：将单个文件追加到 zip.Writer 中
func zipFileToArchive(zw *zip.Writer, srcPath, zipPath string) error {
	file, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer file.Close()

	w, err := zw.Create(zipPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, file)
	return err
}
