<template>
  <div class="content-box">
    <!-- 选择文件与文件夹组合区域 -->
    <div class="select-group">
      <!-- 📄 点击选择文件区域 -->
      <div class="drop-zone" @click="triggerFileSelect">
        <p v-if="!selectedFileName">📄 选择文件（多选）</p>
        <p v-else class="file-name">📄 {{ selectedFileName }}</p>
      </div>

      <!-- 垂直分割线 -->
      <div class="divider"></div>

      <!-- 📁 点击选择文件夹区域 -->
      <div class="drop-zone" @click="triggerDirectorySelect">
        <p v-if="!selectedFolderName">📁 选择文件夹</p>
        <p v-else class="file-name">📁 {{ selectedFolderName }}</p>
      </div>
    </div>

    <!-- 共享链接与二维码 -->
    <div v-if="shareUrl" class="qr-container">
      <qrcode-vue :value="shareUrl" :size="160" level="H" />
      <p class="url-text">{{ shareUrl }}</p>
      <button class="action-btn" @click="copyLink">复制链接</button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import QrcodeVue from 'qrcode.vue';
import { StartSendMode, SelectFile, SelectDirectory } from '../../wailsjs/go/main/App';

const emit = defineEmits(['update-status']);

const selectedFilePaths = ref([]);
const selectedFileName = ref('');
const selectedFolderName = ref('');
const shareUrl = ref('');
const defaultPort = 8080;

const showStatus = (msg, isError = false) => {
  emit('update-status', { msg, isError });
};

// 触发文件选择
const triggerFileSelect = async () => {
  try {
    const paths = await SelectFile();
    if (paths && paths.length > 0) {
      selectedFolderName.value = '';
      selectedFileName.value = paths.length === 1
          ? paths[0].split(/[/\\]/).pop()
          : `已选择 ${paths.length} 个文件`;
      initSendMode(paths);
    }
  } catch (err) {
    showStatus('选择文件失败: ' + err, true);
  }
};

// 触发文件夹选择
const triggerDirectorySelect = async () => {
  try {
    const dirPath = await SelectDirectory();
    if (dirPath) {
      selectedFileName.value = '';
      selectedFolderName.value = dirPath.split(/[/\\]/).pop();
      initSendMode([dirPath]);
    }
  } catch (err) {
    showStatus('选择文件夹失败: ' + err, true);
  }
};

// 启动发送服务
const initSendMode = async (filePaths) => {
  selectedFilePaths.value = filePaths;
  showStatus('正在生成传输链接...', false);

  try {
    const url = await StartSendMode(filePaths, defaultPort);
    shareUrl.value = url;
    showStatus('传输服务已启动，扫描二维码即可下载', false);
  } catch (err) {
    showStatus('启动失败: ' + err, true);
  }
};

// 复制链接
const copyLink = () => {
  if (!shareUrl.value) return;
  navigator.clipboard.writeText(shareUrl.value)
      .then(() => showStatus('链接已复制到剪贴板！', false))
      .catch(() => showStatus('复制失败，请手动复制', true));
};
</script>