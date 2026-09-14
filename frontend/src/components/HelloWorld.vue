<script setup>
import { reactive, onMounted } from 'vue'
// 导入 app.go, ip.go, server.go 中导出的 Go 方法
import { GetLocalIP, StartServer, StopServer } from '../../wailsjs/go/main/App'

const data = reactive({
  localIP: '获取中...',
  isServerRunning: false,
  serverUrl: '',
  port: 8080,
  statusMessage: '准备就绪'
})

// 页面加载时自动获取局域网 IP
onMounted(async () => {
  try {
    data.localIP = await GetLocalIP()
  } catch (err) {
    data.localIP = '获取失败: ' + err
  }
})

// 启动局域网 HTTP 传输服务
async function toggleServer() {
  if (!data.isServerRunning) {
    try {
      // 调用 Go 后端 StartServer 方法
      const url = await StartServer(data.port)
      data.serverUrl = `http://${data.localIP}:${data.port}`
      data.isServerRunning = true
      data.statusMessage = '服务运行中，可在手机端打开 URL 访问'
    } catch (err) {
      data.statusMessage = '启动失败: ' + err
    }
  } else {
    try {
      // 调用 Go 后端 StopServer 方法
      await StopServer()
      data.isServerRunning = false
      data.serverUrl = ''
      data.statusMessage = '服务已停止'
    } catch (err) {
      data.statusMessage = '停止失败: ' + err
    }
  }
}
</script>

<template>
  <main>
    <div class="card">
      <h2>LanShare 文件传输服务</h2>

      <div id="result" class="result">
        <p>本机 IP: <strong>{{ data.localIP }}</strong></p>
        <p v-if="data.isServerRunning">
          访问地址: <a :href="data.serverUrl" target="_blank">{{ data.serverUrl }}</a>
        </p>
        <p class="status">{{ data.statusMessage }}</p>
      </div>

      <div id="input" class="input-box">
        <button
            class="btn"
            :class="{ 'btn-stop': data.isServerRunning }"
            @click="toggleServer"
        >
          {{ data.isServerRunning ? '停止服务' : '启动服务' }}
        </button>
      </div>
    </div>
  </main>
</template>

<style scoped>
.card {
  max-width: 500px;
  margin: 2rem auto;
  padding: 2rem;
  border-radius: 8px;
  background-color: #f9f9f9;
  box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  text-align: center;
}

.result {
  margin: 1.5rem auto;
  line-height: 1.6;
}

.status {
  color: #666;
  font-size: 0.9rem;
}

.input-box .btn {
  width: 120px;
  height: 36px;
  line-height: 36px;
  border-radius: 4px;
  border: none;
  background-color: #4caf50;
  color: white;
  font-weight: bold;
  cursor: pointer;
  transition: background-color 0.2s;
}

.input-box .btn:hover {
  background-color: #45a049;
}

.input-box .btn.btn-stop {
  background-color: #f44336;
}

.input-box .btn.btn-stop:hover {
  background-color: #da190b;
}
</style>