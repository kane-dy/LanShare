package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// StartServer 启动统一服务（同时支持 H5 分享库与 H5 接收上传）
func (a *App) StartServer(port int) (string, error) {
	a.serverLock.Lock()
	defer a.serverLock.Unlock()

	// 1. 如果服务已经启动过，直接获取 IP 返回根目录页面，不再重复启动或关闭
	if a.server != nil {
		localIP, err := a.GetLocalIP()
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("http://%s:%d/", localIP, port), nil
	}

	if err := os.MkdirAll(a.saveDirectory, 0755); err != nil {
		return "", fmt.Errorf("无法创建接收文件夹: %v", err)
	}

	mux := http.NewServeMux()

	// ---------------- 1. 共享库路由 ----------------
	// 根路径 GET / : 响应分享库 H5 页面
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		items := a.GetSharedFiles()
		w.Write([]byte(getShareHubHTML(items)))
	})

	// 单文件强制流式下载路由
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

		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		encodedFileName := url.QueryEscape(fi.Name())
		encodedFileName = strings.ReplaceAll(encodedFileName, "+", "%20")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"; filename*=UTF-8''%s", encodedFileName, encodedFileName))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", fi.Size()))

		_, _ = io.Copy(w, file)
	})

	// 文件移除 API
	mux.HandleFunc("/api/delete", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		filePath := r.URL.Query().Get("path")
		if filePath == "" {
			http.Error(w, "缺少文件路径参数", http.StatusBadRequest)
			return
		}

		err := a.RemoveSharedFile(filePath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// ---------------- 2. 接收文件路由 ----------------
	// GET /upload : 返回上传 H5 页面
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(getUploadHTML()))
	})

	// POST /api/upload : 处理接收上传逻辑
	mux.HandleFunc("/api/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "仅支持 POST 请求", http.StatusMethodNotAllowed)
			return
		}

		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			http.Error(w, fmt.Sprintf("解析表单失败: %v", err), http.StatusBadRequest)
			return
		}

		formData := r.MultipartForm
		files := formData.File["files"]
		if len(files) == 0 {
			http.Error(w, "未选择任何文件", http.StatusBadRequest)
			return
		}

		isShare := r.FormValue("is_share") == "true"

		for _, fileHeader := range files {
			file, err := fileHeader.Open()
			if err != nil {
				continue
			}

			dstPath := filepath.Join(a.saveDirectory, fileHeader.Filename)
			out, err := os.Create(dstPath)
			if err != nil {
				file.Close()
				continue
			}

			_, _ = io.Copy(out, file)
			file.Close()
			out.Close()

			if isShare {
				_ = a.AddSharedFile(SharedItem{
					FileName: fileHeader.Filename,
					FilePath: dstPath,
					FileSize: fileHeader.Size,
				})
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("文件上传成功！"))
	})

	a.server = &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second, // 读取请求超时
		WriteTimeout: 10 * time.Second, // 写入响应超时
		IdleTimeout:  30 * time.Second, // 空闲连接超时（自动回收无效的空闲连接）
	}
	a.currentMode = "combined"

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("统一传输服务异常终止: %v\n", err)
		}
	}()

	localIP, err := a.GetLocalIP()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%d/", localIP, port), nil
}
