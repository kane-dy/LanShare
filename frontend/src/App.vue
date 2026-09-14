<template>
  <div class="container">
    <!-- 头部模式切换 Tab -->
    <div class="tab-header">
      <button
          :class="['tab-btn', activeMode === 'send' ? 'active' : '']"
          @click="switchMode('send')"
      >
        📤 发送文件
      </button>
      <button
          :class="['tab-btn', activeMode === 'receive' ? 'active' : '']"
          @click="switchMode('receive')"
      >
        📥 接收文件
      </button>
    </div>

    <!-- 1. 发送文件模式 -->
    <div v-if="activeMode === 'send'" class="content-box">
      <div class="drop-zone" @click="triggerFileSelect">
        <p v-if="!selectedFileText">点击选择文件（支持多选）</p>
        <p v-else class="file-name">📄 {{ selectedFileText }}</p>
      </div>

      <!-- 共享链接与二维码 -->
      <div v-if="shareUrl" class="qr-container">
        <qrcode-vue :value="shareUrl" :size="160" level="H" />
        <p class="url-text">{{ shareUrl }}</p>
        <button class="action-btn" @click="copyLink">复制链接</button>
      </div>
    </div>

    <!-- 2. 接收文件模式 -->
    <div v-else class="content-box">
      <div v-if="!receiveUrl" class="start-receive-box">
        <p>开启接收服务后，局域网内的其他设备（手机/电脑）扫描二维码或在浏览器打开链接即可上传文件。</p>
        <button class="action-btn primary" @click="startReceive">开启接收服务</button>
      </div>

      <div v-else class="qr-container">
        <qrcode-vue :value="receiveUrl" :size="160" level="H" />
        <p class="url-text">{{ receiveUrl }}</p>
        <div class="btn-group">
          <button class="action-btn" @click="copyLink">复制链接</button>
          <button class="action-btn secondary" @click="openFolder">打开接收文件夹</button>
        </div>
      </div>

      <!-- 保存路径展示与修改设置区域 -->
      <div class="settings-box">
        <div class="path-display">
          <span class="path-label">接收文件路径：</span>
          <input type="text" :value="savePath" class="path-input" readonly />
          <button class="action-btn secondary small-btn" @click="changeDirectory">更改目录</button>
        </div>
      </div>
    </div>

    <!-- 底部状态提示 -->
    <div v-if="statusMsg" class="status-bar" :class="{ error: isError }">
      {{ statusMsg }}
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import QrcodeVue from 'qrcode.vue';

// 从自动生成的 Go 绑定模块中导入 API
import {
  StartSendMode,
  StartReceiveMode,
  StopServer,
  OpenSaveDirectory,
  SelectFile,
  GetSaveDirectory,
  SelectSaveDirectory
} from '../wailsjs/go/main/App';

const activeMode = ref('send'); // 'send' 或 'receive'
const selectedFilePaths = ref([]);
const selectedFileText = ref('');
const shareUrl = ref('');
const receiveUrl = ref('');
const statusMsg = ref('');
const isError = ref(false);
const savePath = ref('');
const defaultPort = 8080;

// 组件挂载时初始化读取当前保存路径
onMounted(async () => {
  try {
    savePath.value = await GetSaveDirectory();
  } catch (err) {
    console.error('获取保存目录失败:', err);
  }
});

// 点击更改保存路径处理逻辑
const changeDirectory = async () => {
  try {
    const newPath = await SelectSaveDirectory();
    if (newPath) {
      savePath.value = newPath;
      showStatus('保存目录修改成功！', false);
    }
  } catch (err) {
    showStatus('修改保存目录失败: ' + err, true);
  }
};

// 切换模式时清理原有的服务与状态
const switchMode = async (mode) => {
  if (activeMode.value === mode) return;
  activeMode.value = mode;
  shareUrl.value = '';
  receiveUrl.value = '';
  selectedFilePaths.value = [];
  selectedFileText.value = '';
  statusMsg.value = '';

  try {
    await StopServer();
  } catch (err) {
    showStatus('停止服务失败: ' + err, true);
  }
};

// 触发文件多选
const triggerFileSelect = async () => {
  try {
    const paths = await SelectFile();
    if (paths && paths.length > 0) {
      initSendMode(paths);
    }
  } catch (err) {
    showStatus('选择文件失败: ' + err, true);
  }
};

// 启动发送服务（支持传入文件路径切片数组）
const initSendMode = async (filePaths) => {
  selectedFilePaths.value = filePaths;

  // 智能展示选择文件的提示文本
  if (filePaths.length === 1) {
    selectedFileText.value = filePaths[0].split(/[/\\]/).pop();
  } else {
    selectedFileText.value = `已选择 ${filePaths.length} 个文件`;
  }

  showStatus('正在生成传输链接...', false);

  try {
    const url = await StartSendMode(filePaths, defaultPort);
    shareUrl.value = url;
    showStatus('传输服务已启动，扫描二维码即可下载', false);
  } catch (err) {
    showStatus('启动失败: ' + err, true);
  }
};

// 启动接收服务
const startReceive = async () => {
  showStatus('正在开启接收服务...', false);
  try {
    const url = await StartReceiveMode(defaultPort);
    receiveUrl.value = url;
    showStatus('接收服务已开启', false);
  } catch (err) {
    showStatus('启动失败: ' + err, true);
  }
};

// 复制链接
const copyLink = () => {
  const targetUrl = shareUrl.value || receiveUrl.value;
  if (!targetUrl) return;

  navigator.clipboard.writeText(targetUrl)
      .then(() => showStatus('链接已复制到剪贴板！', false))
      .catch(() => showStatus('复制失败，请手动复制', true));
};

// 打开本地接收目录
const openFolder = async () => {
  try {
    await OpenSaveDirectory();
  } catch (err) {
    showStatus('打开文件夹失败: ' + err, true);
  }
};

// 状态提示显示控制
const showStatus = (msg, error = false) => {
  statusMsg.value = msg;
  isError.value = error;
};
</script>

<style scoped>
.container {
  padding: 16px;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
  user-select: none;
}

.tab-header {
  display: flex;
  background-color: #f1f5f9;
  padding: 4px;
  border-radius: 8px;
  margin-bottom: 20px;
}

.tab-btn {
  flex: 1;
  padding: 8px 0;
  border: none;
  background: transparent;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  color: #64748b;
  transition: all 0.2s;
}

.tab-btn.active {
  background: white;
  color: #2563eb;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.drop-zone {
  border: 2px dashed #cbd5e1;
  border-radius: 12px;
  padding: 24px;
  text-align: center;
  background: #f8fafc;
  cursor: pointer;
  color: #64748b;
  margin-bottom: 16px;
}

.drop-zone:hover {
  border-color: #2563eb;
  background: #eff6ff;
}

.file-name {
  color: #0f172a;
  font-weight: bold;
  word-break: break-all;
}

.qr-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-top: 16px;
}

.url-text {
  font-size: 0.85rem;
  color: #475569;
  background: #f1f5f9;
  padding: 6px 12px;
  border-radius: 4px;
  margin: 12px 0;
  word-break: break-all;
}

.action-btn {
  background: #2563eb;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.9rem;
}

.action-btn.secondary {
  background: #e2e8f0;
  color: #334155;
}

.btn-group {
  display: flex;
  gap: 8px;
}

.start-receive-box {
  text-align: center;
  padding: 20px 0;
  color: #64748b;
  line-height: 1.6;
}

.settings-box {
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid #e2e8f0;
}

.path-display {
  display: flex;
  align-items: center;
  gap: 8px;
}

.path-label {
  font-size: 0.85rem;
  color: #475569;
  white-space: nowrap;
}

.path-input {
  flex: 1;
  padding: 6px 10px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  background-color: #f8fafc;
  color: #334155;
  font-size: 0.85rem;
  outline: none;
}

.small-btn {
  padding: 6px 12px;
  font-size: 0.8rem;
  white-space: nowrap;
}

.status-bar {
  margin-top: 16px;
  padding: 8px 12px;
  border-radius: 6px;
  background-color: #f0fdf4;
  color: #166534;
  font-size: 0.85rem;
  text-align: center;
}

.status-bar.error {
  background-color: #fef2f2;
  color: #991b1b;
}
</style>