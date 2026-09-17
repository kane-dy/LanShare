package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// StartReceiveMode 开启收文件模式（生成上传服务与网页）
func (a *App) StartReceiveMode(port int) (string, error) {
	_ = a.StopServer()

	a.serverLock.Lock()
	defer a.serverLock.Unlock()

	if err := os.MkdirAll(a.saveDirectory, 0755); err != nil {
		return "", fmt.Errorf("无法创建接收文件夹: %v", err)
	}

	mux := http.NewServeMux()

	// GET / : 响应内置 H5 上传网页
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(getUploadHTML()))
	})

	// POST /upload : 处理文件上传（流式写入零内存溢出）
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "仅支持 POST 请求", http.StatusMethodNotAllowed)
			return
		}

		// 限制内存使用（最大 32MB 内存缓冲，超出写入临时文件）
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

		// 🌟 1. 获取前端发来的“是否同时分享”复选框状态
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

			// 使用 io.Copy 流式写入，防止大文件吃满电脑内存
			_, _ = io.Copy(out, file)

			file.Close()
			out.Close()

			// 🌟 2. 若用户勾选了“同时添加到局域网分享库”，将接收到的文件写入分享记录
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
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	a.currentMode = "receive"

	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("收文件服务异常终止: %v\n", err)
		}
	}()

	localIP, err := a.GetLocalIP()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("http://%s:%d/", localIP, port), nil
}
