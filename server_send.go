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

	var downloadFileName string
	var isSingleFile bool = false
	var singleFilePath string

	if len(filePaths) == 1 {
		info, err := os.Stat(filePaths[0])
		if err != nil {
			return "", fmt.Errorf("文件不存在: %v", err)
		}
		if !info.IsDir() {
			isSingleFile = true
			singleFilePath = filePaths[0]
			downloadFileName = info.Name()
		}
	}

	if !isSingleFile {
		downloadFileName = fmt.Sprintf("LanShare_传输文件_%s.zip", time.Now().Format("150405"))
	}

	mux := http.NewServeMux()

	ext := filepath.Ext(downloadFileName)
	if ext == "" && !isSingleFile {
		ext = ".zip"
	}
	cleanRoute := fmt.Sprintf("/download/file%s", ext)

	// 注册临时传输下载路由
	mux.HandleFunc("/download/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		encodedFileName := url.QueryEscape(downloadFileName)
		contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", encodedFileName, encodedFileName)
		w.Header().Set("Content-Disposition", contentDisposition)

		if isSingleFile {
			file, err := os.Open(singleFilePath)
			if err != nil {
				http.Error(w, "无法读取文件", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			fi, err := file.Stat()
			if err != nil {
				http.Error(w, "获取文件状态失败", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Accept-Ranges", "bytes")
			http.ServeContent(w, r, downloadFileName, fi.ModTime(), file)
			return
		}

		zw := zip.NewWriter(w)
		defer zw.Close()

		for _, path := range filePaths {
			info, err := os.Stat(path)
			if err != nil {
				continue
			}

			if info.IsDir() {
				baseDir := filepath.Dir(path)
				_ = filepath.Walk(path, func(filePath string, fi os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					relPath, err := filepath.Rel(baseDir, filePath)
					if err != nil {
						return err
					}
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
				_ = zipFileToArchive(zw, path, info.Name())
			}
		}
	})

	// 注册分享库全局下载路由
	mux.HandleFunc("/download/shared", func(w http.ResponseWriter, r *http.Request) {
		filePath := r.URL.Query().Get("path")
		if filePath == "" {
			http.Error(w, "未指定文件路径", http.StatusBadRequest)
			return
		}

		file, err := os.Open(filePath)
		if err != nil {
			http.Error(w, "文件不存在或已被删除", http.StatusNotFound)
			return
		}
		defer file.Close()

		fi, err := file.Stat()
		if err != nil {
			http.Error(w, "无法获取文件信息", http.StatusInternalServerError)
			return
		}

		encodedFileName := url.QueryEscape(fi.Name())
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", encodedFileName, encodedFileName))
		w.Header().Set("Accept-Ranges", "bytes")

		http.ServeContent(w, r, fi.Name(), fi.ModTime(), file)
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

// StartShareHubMode 启动全局分享库模式（扫码后进入文件表格列表）
// StartShareHubMode 启动全局分享库模式（扫码后进入文件表格列表）
func (a *App) StartShareHubMode(port int) (string, error) {
	_ = a.StopServer()

	a.serverLock.Lock()
	defer a.serverLock.Unlock()

	mux := http.NewServeMux()

	// 1. 根路径 GET / : 响应包含文件表格的 H5 网页
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// 动态获取当前的共享文件列表并渲染页面
		items := a.GetSharedFiles()
		w.Write([]byte(getShareHubHTML(items)))
	})

	// 2. 单文件强制下载接口（对标 StartSendMode，将文件名挂载在 URL 路径上）
	mux.HandleFunc("/download/shared/", func(w http.ResponseWriter, r *http.Request) {
		filePath := r.URL.Query().Get("path")
		if filePath == "" {
			http.Error(w, "未指定文件路径", http.StatusBadRequest)
			return
		}

		file, err := os.Open(filePath)
		if err != nil {
			http.Error(w, "文件不存在或已被删除", http.StatusNotFound)
			return
		}
		defer file.Close()

		fi, err := file.Stat()
		if err != nil {
			http.Error(w, "无法获取文件信息", http.StatusInternalServerError)
			return
		}

		// 🌟 防缓存与类型重写策略（与 StartSendMode 一致）
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// 🌟 强行设为附件下载，并处理文件名转码（空格替换为 %20）
		encodedFileName := url.QueryEscape(fi.Name())
		encodedFileName = strings.ReplaceAll(encodedFileName, "+", "%20")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", encodedFileName, encodedFileName))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fi.Size()))

		// 🌟 改用 io.Copy 替代 http.ServeContent，防止 docx/pdf 等文件被浏览器解析在线打开
		_, _ = io.Copy(w, file)
	})

	a.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	a.currentMode = "share"

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("分享库服务异常终止: %v\n", err)
		}
	}()

	localIP, err := a.GetLocalIP()
	if err != nil {
		return "", err
	}

	// 二维码指向根路径地址
	return fmt.Sprintf("http://%s:%d/", localIP, port), nil
}
