package main

import (
	"fmt"
	"net/url"
)

// getUploadHTML 返回收文件模式下给手机/电脑浏览器使用的极简上传界面
func getUploadHTML() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>LanShare 文件上传</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; padding: 20px; background: #f8fafc; color: #334155; margin: 0; }
    .card { max-width: 480px; margin: 40px auto; background: white; padding: 24px; border-radius: 16px; box-shadow: 0 4px 12px rgba(0,0,0,0.05); text-align: center; }
    .drop-box { border: 2px dashed #cbd5e1; border-radius: 12px; padding: 30px 15px; margin: 20px 0; background: #f8fafc; cursor: pointer; }
    .drop-box.hover { border-color: #2563eb; background: #eff6ff; }
    input[type="file"] { display: none; }
    
    /* 复选框样式 */
    .checkbox-container {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
      margin-bottom: 20px;
      font-size: 0.95rem;
      color: #475569;
      cursor: pointer;
    }
    .checkbox-container input[type="checkbox"] {
      width: 18px;
      height: 18px;
      cursor: pointer;
      accent-color: #2563eb;
    }

    .btn { background: #2563eb; color: white; border: none; padding: 12px 24px; border-radius: 8px; font-size: 1rem; width: 100%; cursor: pointer; }
    .btn:disabled { background: #94a3b8; }
    #status { margin-top: 15px; font-size: 0.9rem; font-weight: bold; }
  </style>
</head>
<body>
  <div class="card">
    <h2>📤 上传文件到电脑</h2>
    <div class="drop-box" id="dropZone" onclick="document.getElementById('fileInput').click()">
      <p id="hint">点击选择文件</p>
      <input type="file" id="fileInput" multiple onchange="handleSelect(this.files)">
    </div>

    <!-- 新增：是否分享复选框 -->
    <label class="checkbox-container">
      <input type="checkbox" id="shareCheckbox">
      <span>同时添加到电脑的局域网分享库</span>
    </label>

    <button class="btn" id="uploadBtn" onclick="upload()" disabled>开始上传</button>
    <div id="status"></div>
  </div>

  <script>
    let selectedFiles = [];
    const dropZone = document.getElementById('dropZone');
    const hint = document.getElementById('hint');
    const btn = document.getElementById('uploadBtn');
    const status = document.getElementById('status');
    const shareCheckbox = document.getElementById('shareCheckbox');

    dropZone.addEventListener('dragover', (e) => { e.preventDefault(); dropZone.classList.add('hover'); });
    dropZone.addEventListener('dragleave', () => dropZone.classList.remove('hover'));
    dropZone.addEventListener('drop', (e) => {
      e.preventDefault();
      dropZone.classList.remove('hover');
      handleSelect(e.dataTransfer.files);
    });

    function handleSelect(files) {
      if (files.length > 0) {
        selectedFiles = files;
        hint.innerText = "已选择 " + files.length + " 个文件";
        btn.disabled = false;
      }
    }

    async function upload() {
      if (selectedFiles.length === 0) return;
      const formData = new FormData();
      for (let i = 0; i < selectedFiles.length; i++) {
        formData.append('files', selectedFiles[i]);
      }

      // 新增：向 FormData 添加是否分享字段
      formData.append('is_share', shareCheckbox.checked ? 'true' : 'false');

      btn.disabled = true;
      status.style.color = '#2563eb';
      status.innerText = '正在上传中...';

      try {
        const res = await fetch('/upload', { method: 'POST', body: formData });
        if (res.ok) {
          status.style.color = '#16a34a';
          status.innerText = '🎉 上传成功！';
          selectedFiles = [];
          hint.innerText = '点击选择文件';
          shareCheckbox.checked = false; // 上传成功后重置复选框
        } else {
          const errorMsg = await res.text();
          throw new Error(errorMsg || ('服务器返回错误 (' + res.status + ')'));
        }
      } catch (err) {
        status.style.color = '#dc2626';
        
        if (err.name === 'TypeError' || err.message.includes('fetch') || err.message.includes('NetworkError')) {
          status.innerText = '❌ 无法连接到电脑，传输服务可能已停止或网络已断开';
        } else {
          status.innerText = '❌ 上传失败: ' + err.message;
        }
        
        btn.disabled = false;
      }
    }
  </script>
</body>
</html>`
}

// getShareHubHTML 返回手机/浏览器扫码后看到的共享文件列表表格 H5 页面
func getShareHubHTML(items []SharedItem) string {
	rows := ""
	for _, item := range items {
		// 格式化文件大小
		sizeMB := float64(item.FileSize) / 1024 / 1024
		sizeStr := fmt.Sprintf("%.2f MB", sizeMB)
		if item.FileSize < 1024*1024 {
			sizeStr = fmt.Sprintf("%.2f KB", float64(item.FileSize)/1024)
		}

		// 🌟 修改点：将文件名挂载到路径后（如 /download/shared/文档.docx?path=...）
		// 浏览器识别到真实文件后缀后，会直接触发下载而非在线预览
		downloadUrl := fmt.Sprintf("/download/shared/%s?path=%s", url.PathEscape(item.FileName), url.QueryEscape(item.FilePath))

		rows += fmt.Sprintf(`
       <tr>
          <td class="file-name">%s</td>
          <td class="file-size">%s</td>
          <td><a href="%s" class="dl-btn" download="%s">下载</a></td>
       </tr>`, item.FileName, sizeStr, downloadUrl, item.FileName)
	}

	if len(items) == 0 {
		rows = `<tr><td colspan="3" style="text-align:center;color:#94a3b8;padding:20px;">暂无共享文件</td></tr>`
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>LanShare 局域网分享库</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; padding: 16px; background: #f8fafc; color: #334155; margin: 0; }
    .card { max-width: 600px; margin: 20px auto; background: white; padding: 20px; border-radius: 16px; box-shadow: 0 4px 12px rgba(0,0,0,0.05); }
    h2 { margin-top: 0; color: #0f172a; font-size: 1.25rem; display: flex; align-items: center; gap: 8px; }
    table { width: 100%%; border-collapse: collapse; margin-top: 16px; font-size: 0.9rem; }
    th, td { padding: 12px 8px; text-align: left; border-bottom: 1px solid #f1f5f9; }
    th { background: #f8fafc; color: #64748b; font-weight: 600; }
    .file-name { max-width: 180px; word-break: break-all; font-weight: 500; }
    .file-size { color: #64748b; font-size: 0.85rem; white-space: nowrap; }
    .dl-btn { display: inline-block; background: #2563eb; color: white; text-decoration: none; padding: 6px 14px; border-radius: 6px; font-size: 0.85rem; font-weight: 500; }
    .dl-btn:active { background: #1d4ed8; }
  </style>
</head>
<body>
  <div class="card">
    <h2>📁 局域网共享文件列表</h2>
    <table>
      <thead>
        <tr>
          <th>文件名</th>
          <th>大小</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        %s
      </tbody>
    </table>
  </div>
</body>
</html>`, rows)
}
