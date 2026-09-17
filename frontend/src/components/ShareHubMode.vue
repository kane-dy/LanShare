<template>
  <div class="content-box">
    <!-- 顶部添加/管理按钮 -->
    <div class="select-group">
      <div class="drop-zone" @click="addFilesToHub">
        <p>➕ 点击添加文件到分享库</p>
      </div>
    </div>

    <!-- 核心：界面直接展示二维码 -->
    <div v-if="shareUrl" class="qr-container">
      <qrcode-vue :value="shareUrl" :size="160" level="H" />
      <p class="url-text">{{ shareUrl }}</p>
      <div class="btn-group">
        <button class="action-btn" @click="copyLink">复制链接</button>
        <button class="action-btn secondary" @click="clearHub">清空分享库</button>
      </div>
      <p class="tip-text">📱 手机扫码或浏览器访问此链接，即可查看分享文件表格并下载</p>
    </div>

    <!-- 暂无文件时的提示 -->
    <div v-else class="empty-tip">
      <p>分享库暂无文件</p>
      <p style="font-size: 0.75rem; color: #94a3b8;">添加文件后将自动生成局域网分享二维码</p>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue';
import QrcodeVue from 'qrcode.vue';
import {
  SelectFile,
  AddSharedFile,
  GetSharedFiles,
  DeleteSharedFile,
  StartShareHubMode
} from '../../wailsjs/go/main/App';

const emit = defineEmits(['update-status']);
const shareUrl = ref('');
const defaultPort = 8080;

const showStatus = (msg, isError = false) => {
  emit('update-status', { msg, isError });
};

// 刷新并启动全局分享库服务
const refreshShareHub = async () => {
  try {
    const list = await GetSharedFiles();
    if (list && list.length > 0) {
      // 只要有文件，就启动服务并显示二维码
      const url = await StartShareHubMode(defaultPort);
      shareUrl.value = url;
    } else {
      shareUrl.value = '';
    }
  } catch (err) {
    showStatus('启动分享服务失败: ' + err, true);
  }
};

onMounted(() => {
  refreshShareHub();
});

// 添加文件到分享库
const addFilesToHub = async () => {
  try {
    const paths = await SelectFile();
    if (paths && paths.length > 0) {
      for (const p of paths) {
        const fileName = p.split(/[/\\]/).pop();
        await AddSharedFile({
          file_name: fileName,
          file_path: p,
          file_size: 0
        });
      }
      showStatus('文件已加入分享库！', false);
      await refreshShareHub();
    }
  } catch (err) {
    showStatus('添加文件失败: ' + err, true);
  }
};

// 清空分享库
const clearHub = async () => {
  try {
    const list = await GetSharedFiles();
    for (const item of list) {
      await DeleteSharedFile(item.id);
    }
    showStatus('分享库已清空', false);
    await refreshShareHub();
  } catch (err) {
    showStatus('清空失败: ' + err, true);
  }
};

// 复制链接
const copyLink = () => {
  if (!shareUrl.value) return;
  navigator.clipboard.writeText(shareUrl.value)
      .then(() => showStatus('共享链接已复制！', false))
      .catch(() => showStatus('复制失败', true));
};
</script>

<style scoped>
.tip-text {
  margin-top: 14px;
  font-size: 0.8rem;
  color: #64748b;
  text-align: center;
}

.empty-tip {
  text-align: center;
  color: #64748b;
  padding: 40px 0;
  line-height: 1.6;
}
</style>