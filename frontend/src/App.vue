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
      <!-- 🌟 新增分享库 Tab -->
      <button
          :class="['tab-btn', activeMode === 'share' ? 'active' : '']"
          @click="switchMode('share')"
      >
        🔗 分享库
      </button>
    </div>

    <!-- 动态组件渲染区域 -->
    <SendMode v-if="activeMode === 'send'" @update-status="handleStatusUpdate" />
    <ReceiveMode v-else-if="activeMode === 'receive'" @update-status="handleStatusUpdate" />
    <!-- 🌟 渲染 ShareHubMode -->
    <ShareHubMode v-else-if="activeMode === 'share'" @update-status="handleStatusUpdate" />

    <!-- 底部状态提示 -->
    <div v-if="statusMsg" class="status-bar" :class="{ error: isError }">
      {{ statusMsg }}
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue';
import SendMode from './components/SendMode.vue';
import ReceiveMode from './components/ReceiveMode.vue';
import ShareHubMode from './components/ShareHubMode.vue'; // 🌟 引入新组件
// import { StopServer } from '../wailsjs/go/main/App';

const activeMode = ref('send');
const statusMsg = ref('');
const isError = ref(false);

const handleStatusUpdate = ({ msg, isError: errStatus }) => {
  statusMsg.value = msg;
  isError.value = errStatus;
};

// 切换模式时清理原有的服务与状态
const switchMode = (mode) => {
  if (activeMode.value === mode) return;
  activeMode.value = mode;
  statusMsg.value = '';
};

</script>

<style>
/* 引入公共 CSS 文件 */
@import './assets/style.css';
</style>