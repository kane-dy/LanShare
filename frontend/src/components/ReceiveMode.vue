<template>
  <div class="content-box">
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
</template>

<script setup>
import { ref, onMounted } from 'vue';
import QrcodeVue from 'qrcode.vue';
import {
  StartReceiveMode,
  OpenSaveDirectory,
  GetSaveDirectory,
  SelectSaveDirectory
} from '../../wailsjs/go/main/App';

const emit = defineEmits(['update-status']);

const receiveUrl = ref('');
const savePath = ref('');
const defaultPort = 8080;

const showStatus = (msg, isError = false) => {
  emit('update-status', { msg, isError });
};

onMounted(async () => {
  try {
    savePath.value = await GetSaveDirectory();
  } catch (err) {
    console.error('获取保存目录失败:', err);
  }
});

// 更改保存路径
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

// 开启接收服务
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
  if (!receiveUrl.value) return;
  navigator.clipboard.writeText(receiveUrl.value)
      .then(() => showStatus('链接已复制到剪贴板！', false))
      .catch(() => showStatus('复制失败，请手动复制', true));
};

// 打开文件夹
const openFolder = async () => {
  try {
    await OpenSaveDirectory();
  } catch (err) {
    showStatus('打开文件夹失败: ' + err, true);
  }
};
</script>

<style scoped>
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
</style>