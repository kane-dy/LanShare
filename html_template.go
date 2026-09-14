package main

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
    <button class="btn" id="uploadBtn" onclick="upload()" disabled>开始上传</button>
    <div id="status"></div>
  </div>

  <script>
    let selectedFiles = [];
    const dropZone = document.getElementById('dropZone');
    const hint = document.getElementById('hint');
    const btn = document.getElementById('uploadBtn');
    const status = document.getElementById('status');

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
        } else {
          // 读取服务端返回的错误信息（如磁盘空间不足、权限问题等）
          const errorMsg = await res.text();
          throw new Error(errorMsg || ('服务器返回错误 (' + res.status + ')'));
        }
      } catch (err) {
        status.style.color = '#dc2626';
        
        // 当服务停止或断网时，fetch 会抛出 TypeError 或网络相关异常
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
