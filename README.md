# 📁 LanShare — 项目开发文档

> 局域网文件传输工具 · Wails v2 (Go + Vue 3)

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![Vue](https://img.shields.io/badge/Vue-3.5+-42b883?logo=vue.js)](https://vuejs.org)
[![Wails](https://img.shields.io/badge/Wails-v2-6b7280)](https://wails.io)

---

## 1. 项目概述

**LanShare** 是一款面向**同一局域网**的轻量级文件传输桌面应用。核心思路：在本机启动一个轻量 HTTP 服务，自动生成本地域 URL 和二维码，接收方（手机/电脑）用浏览器扫码即可完成下载或上传——**零安装、零账号、零配置**。

两种核心模式：

| 模式 | 行为 | 典型场景 |
|------|------|----------|
| 📤 **发送模式** | 把本机文件作为 HTTP 下载资源暴露出去 | 电脑 → 手机发文件 |
| 📥 **接收模式** | 在本机启动 HTTP 上传服务（含内嵌 H5 上传页） | 手机 → 电脑收文件 |

---

## 2. 技术架构

### 2.1 整体分层

```
┌──────────────────────────────────────────────────────────────┐
│                   Wails v2 桌面容器                           │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  前端层 (frontend/src/App.vue)                         │  │
│  │  ┌────────────┐  ┌──────────────┐  ┌───────────────┐  │  │
│  │  │ Tab 切换    │  │ 二维码渲染    │  │ 状态提示/复制  │  │  │
│  │  └────────────┘  └──────────────┘  └───────────────┘  │  │
│  └──────────────────────────┬─────────────────────────────┘  │
│                             │ window.go.main.App.*            │
│  ┌──────────────────────────▼─────────────────────────────┐  │
│  │  绑定层 (wailsjs/go/main/App.js) — 自动生成，勿手动改    │  │
│  └──────────────────────────┬─────────────────────────────┘  │
│                             │                                  │
│  ┌──────────────────────────▼─────────────────────────────┐  │
│  │  后端层 (LanShare/*.go)                                │  │
│  │  ┌─────────────┐ ┌──────────────┐ ┌─────────────────┐ │  │
│  │  │ 发送 HTTP   │ │ 接收 HTTP    │ │ 配置/网络/系统   │ │  │
│  │  │ server_send │ │ server_recv  │ │ app/network/html │ │  │
│  │  └─────────────┘ └──────────────┘ └─────────────────┘ │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
           ↕ HTTP :8080
    ┌────────────────────────┐
    │  局域网内任意浏览器      │
    └────────────────────────┘
```

### 2.2 技术栈

| 层级 | 技术 | 版本 | 备注 |
|------|------|------|------|
| 桌面框架 | Wails v2 | v2.15.0 | Go + WebView2 |
| 后端 | Go | 1.25+ | 标准库 `net/http` / `archive/zip` / `os` |
| 前端 | Vue 3 | ^3.5.0 | Composition API |
| 构建 | Vite | ^7.0.0 | 前端打包 |
| 二维码 | qrcode.vue | ^3.11.0 | URL → 二维码 Canvas |
| 打包 | Go embed | 1.16+ | 前端资源嵌入二进制 |

### 2.3 为什么选 Wails 而不是 Electron/Tauri？

- **比 Electron 轻**：Go + WebView2，二进制小、内存占用低
- **比 Tauri 成熟**：Go 生态比 Rust 更贴合本项目的 HTTP 服务需求
- **前后端通信简单**：`Bind` 机制自动把 Go 方法暴露给 JS，无需手写 IPC

---

## 3. 目录结构

```
LanShare/
│
├── main.go                    程序入口 (48 行)
│   └── //go:embed all:frontend/dist   ← 前端资源打包进二进制
│
├── app.go                      核心 App 结构体 + 配置 (173 行)
│   ├── Config 结构体          ← JSON 持久化模型
│   ├── App 结构体             ← 全局服务状态
│   ├── NewApp()               ← 构造函数（加载配置、初始化路径）
│   ├── loadConfig / saveConfig ← JSON 读写
│   ├── startUp()              ← Wails 生命周期钩子
│   └── GetSaveDirectory / SelectSaveDirectory / SetSaveDirectory
│       StopServer
│
├── server_send.go             发送模式 (167 行)
│   ├── StartSendMode(filePaths, port)  ← 核心入口
│   └── zipFileToArchive(zw, srcPath, zipPath)  ← ZIP 打包辅助
│
├── server_receive.go          接收模式 (97 行)
│   └── StartReceiveMode(port)          ← 核心入口
│
├── network.go                 网络/系统工具 (112 行)
│   ├── GetLocalIP()           ← 双策略 IP 获取
│   ├── OpenSaveDirectory()    ← 跨平台打开文件夹
│   └── SelectFile()           ← 系统多选文件对话框
│
├── html_template.go           内嵌 H5 上传页 (95 行)
│   └── getUploadHTML()        ← 返回完整 HTML 字符串
│
├── go.mod / go.sum            Go 依赖锁定
│
├── frontend/
│   ├── src/App.vue            主界面 (351 行)
│   ├── wailsjs/go/main/App.js Wails 自动生成的 Go↔JS 代理
│   ├── wailsjs/runtime/       Wails JS 运行时
│   ├── vite.config.js         Vite 配置
│   ├── package.json
│   └── dist/                  构建产物 → 被 go:embed 打包
│
└── build/                     打包资源 (图标、NSIS 安装脚本、plist)
```

**总代码量（不含前端 node_modules 和 build 产物）：** 6 个 Go 文件 ≈ 692 行 + 1 个 Vue 组件 ≈ 351 行

---

## 4. 核心数据结构

### 4.1 App 结构体（`app.go:22-29`）

```go
type App struct {
    ctx           context.Context  // Wails 生命周期上下文，由 startUp 注入
    server        *http.Server     // 当前运行的 HTTP 服务实例
    serverLock    sync.Mutex       // 保护 server 字段的并发锁
    saveDirectory string           // 接收文件的保存目录（可配置）
    configPath    string           // 配置文件绝对路径（%AppData%/LanShare/config.json）
    currentMode   string           // 当前模式："send" | "receive" | ""
}
```

字段说明：

| 字段 | 类型 | 初始值 | 说明 |
|------|------|--------|------|
| `ctx` | `context.Context` | `nil` | Wails `OnStartup` 回调时注入 |
| `server` | `*http.Server` | `nil` | 每次 Start 时创建，Stop 后置 `nil` |
| `serverLock` | `sync.Mutex` | — | 保证快速切换模式时不会竞态 |
| `saveDirectory` | `string` | `~/Downloads/LanShareFile` | 用户修改后持久化到 config.json |
| `configPath` | `string` | `%AppData%/LanShare/config.json` | 由 `getConfigFilePath()` 动态生成 |
| `currentMode` | `string` | `""` | 辅助标记，目前未做复杂判断 |

### 4.2 Config 结构体（`app.go:17-19`）

```go
type Config struct {
    SaveDirectory string `json:"save_directory"`
}
```

序列化示例：
```json
{
  "save_directory": "C:\\Users\\Alice\\Downloads\\LanShareFile"
}
```

---

## 5. 发送模式详解

### 5.1 入口：`StartSendMode(filePaths []string, port int) (string, error)`

完整调用链：

```
StartSendMode([]string{"D:/照片", "D:/报告.pdf"}, 8080)
│
├── a.StopServer()                     // 先停掉旧服务（2s 超时优雅关闭）
├── a.serverLock.Lock() / defer Unlock()  // 加锁
│
├── 判断文件类型
│   ├── len(filePaths) == 1 && !info.IsDir()
│   │   └── isSingleFile = true
│   │       downloadFileName = info.Name()
│   └── 其他情况（多文件 或 含文件夹）
│       └── downloadFileName = "LanShare_传输文件_150405.zip"
│
├── 创建 http.NewServeMux()
│   └── 注册路由 "/download/"
│       ├── 设置防缓存 Header (Cache-Control, Pragma, Expires)
│       ├── 设置 Content-Type: application/octet-stream
│       ├── 设置 Content-Disposition（UTF-8 编码文件名）
│       │
│       ├── 情况 A：单文件
│       │   └── os.Open → file.Stat → http.ServeContent（支持 Range/断点续传）
│       │
│       └── 情况 B：多文件/文件夹
│           └── zip.NewWriter(w)  ← 直接写 HTTP ResponseWriter，边压边传
│               ├── 普通文件 → zipFileToArchive()
│               └── 文件夹 → filepath.Walk 递归
│                   ├── 目录 → zw.CreateHeader（只写元数据，不写内容）
│                   └── 文件 → zipFileToArchive()
│                       ├── os.Open(srcPath)
│                       ├── zw.Create(zipPath) → 返回 io.Writer
│                       └── io.Copy(zwWriter, file)
│
├── a.server = &http.Server{Addr: ":8080", Handler: mux}
├── a.currentMode = "send"
├── go a.server.ListenAndServe()       // 异步启动
├── a.GetLocalIP()                     // 获取局域网 IP
└── 返回 URL: "http://192.168.1.100:8080/download/file.zip?t=1726..."
```

### 5.2 下载路由的 HTTP Header 策略

| Header | 值 | 目的 |
|--------|-----|------|
| `Content-Type` | `application/octet-stream` | 强制浏览器走下载流程 |
| `X-Content-Type-Options` | `nosniff` | 禁止 MIME 嗅探 |
| `Content-Disposition` | `attachment; filename="..."; filename*=UTF-8''...` | 正确处理中文文件名（RFC 5987） |
| `Cache-Control` | `no-cache, no-store, must-revalidate` | 防浏览器缓存，避免下次打开看到旧文件 |
| `Content-Length` | 单文件时设置 | 让浏览器显示下载进度 |

### 5.3 单文件 vs 多文件的实现差异

| 维度 | 单文件 | 多文件/文件夹 |
|------|--------|---------------|
| 文件名 | 原始文件名（如 `报告.pdf`） | `LanShare_传输文件_HHMMSS.zip` |
| 传输方式 | `http.ServeContent`（支持 Range 续传） | `archive/zip` 流式打包写入 Response |
| 内存占用 | 极低（os.File.ReadAt 内核缓冲） | 极低（边压边写，不攒内存） |
| Content-Length | 可预知，设 Header | 无法预知，浏览器 chunked 传输 |
| 取消支持 | Range 自动续传 | 浏览器中断即停止，不支持续传 |

---

## 6. 接收模式详解

### 6.1 入口：`StartReceiveMode(port int) (string, error)`

```
StartReceiveMode(8080)
│
├── a.StopServer()
├── a.serverLock.Lock() / defer Unlock()
├── os.MkdirAll(a.saveDirectory)       // 确保保存目录存在
│
├── 创建 http.NewServeMux()
│   ├── GET  "/"  → getUploadHTML()    // 返回内嵌 H5 上传页面
│   └── POST "/upload" → 处理 multipart/form-data
│       ├── r.ParseMultipartForm(32MB) // 32MB 内存缓冲，超出落临时文件
│       ├── formData.File["files"]     // 取所有文件
│       └── 遍历每个 fileHeader
│           ├── fileHeader.Open() → io.ReadCloser
│           ├── os.Create(filepath.Join(saveDir, filename)) → *os.File
│           └── io.Copy(out, file)    // 流式写入，零内存溢出
│
├── a.server = &http.Server{Addr: ":8080", Handler: mux}
├── a.currentMode = "receive"
├── go a.server.ListenAndServe()
├── a.GetLocalIP()
└── 返回 URL: "http://192.168.1.100:8080/"
```

### 6.2 内嵌 H5 上传页（`html_template.go`）

接收方浏览器访问 `GET /` 时，后端直接返回一个 95 行的单文件 HTML：

| 特性 | 实现方式 |
|------|----------|
| 拖拽上传 | `dragover` / `dragleave` / `drop` 事件监听 |
| 点击选择 | `<input type="file" multiple>` |
| 多文件 | `FormData.append('files', each)` 循环 |
| 上传请求 | `fetch('/upload', { method: 'POST', body: formData })` |
| 错误处理 | 区分服务端返回错误 vs 网络断连（TypeError） |
| 移动端 | `viewport` meta + 大按钮 + `-apple-system` 字体栈 |

**注意**：H5 页面不支持**显示上传进度** — 后端也没有做 SSE 或 WebSocket 推送。如果未来需要进度，需要改造前后端。

---

## 7. 网络 IP 获取策略

`GetLocalIP()`（`network.go:15-51`）采用**双策略降级**：

```
策略 1：UDP 伪连接
  net.Dial("udp", "8.8.8.8:80")
  → 让 OS 路由表选择最佳出向网卡
  → conn.LocalAddr().(*net.UDPAddr).IP.String()
  ✅ 有外网时最准确（VPN、多网卡场景也能识别）
  ⚠️ 需要能连上 8.8.8.8

失败 → 策略 2：遍历网卡
  net.Interfaces()
  → 过滤 !FlagUp 或 FlagLoopback
  → iface.Addrs() → 取第一个 IPv4
  ✅ 完全离线也能用
  ⚠️ 多网卡时取第一个，不一定对
```

为什么不用 `net.Listen` 监听 `0.0.0.0` 来反推？——因为那会启动实际监听，在只想查询 IP 时显得重了。UDP 伪连接既轻量又准确。

---

## 8. 配置持久化

### 8.1 存储位置

| 平台 | 路径 |
|------|------|
| Windows | `%AppData%\LanShare\config.json` |
| macOS | `~/Library/Application Support/LanShare/config.json` |
| Linux | `~/.config/LanShare/config.json` |

由 `os.UserConfigDir()` 自动推导（`app.go:32-40`）。

### 8.2 读写时序

```
程序启动
│
├── NewApp()
│   ├── savePath = "~/Downloads/LanShareFile"  // 默认值
│   ├── userDir = os.UserHomeDir()
│   │   └── savePath = userDir + "/Downloads/LanShareFile"  // 覆盖默认
│   └── app.loadConfig()                          // 读取 JSON
│       └── cfg.SaveDirectory != "" → 再次覆盖
│
├── startUp(ctx)
│   └── os.MkdirAll(a.saveDirectory, 0755)       // 确保目录存在
│
└── 用户点击"更改目录"
    ├── SelectSaveDirectory() → runtime.OpenDirectoryDialog
    ├── a.saveDirectory = selection
    └── a.saveConfig() → json.MarshalIndent → os.WriteFile(configPath)
```

**三级 fallback 机制**：默认值 → 用户主目录 → 历史配置文件。

---

## 9. HTTP API 参考

### 9.1 发送模式路由

| 方法 | 路径 | 响应 | 说明 |
|------|------|------|------|
| `GET` | `/download/file.<ext>` | 文件流 / ZIP 流 | `<ext>` 由文件名动态决定（`.pdf` / `.zip` 等） |
| `HEAD` | 同上 | Header only | Go 默认支持，客户端探测用 |

### 9.2 接收模式路由

| 方法 | 路径 | Content-Type | 响应 | 说明 |
|------|------|--------------|------|------|
| `GET` | `/` | `text/html; charset=utf-8` | 内嵌 H5 页面 | 非 `/` 路径返回 404 |
| `POST` | `/upload` | `multipart/form-data` | 纯文本 `"文件上传成功！"` | 表单字段名 `files`，支持多文件 |

### 9.3 Go → JS 绑定方法（按功能分组）

#### 服务控制

| 方法 | 签名 | 说明 |
|------|------|------|
| `StartSendMode` | `(filePaths []string, port int) → (string, error)` | 启动发送服务，返回下载 URL |
| `StartReceiveMode` | `(port int) → (string, error)` | 启动接收服务，返回上传页 URL |
| `StopServer` | `() → error` | 优雅关闭 HTTP 服务（2s 超时） |

#### 文件系统

| 方法 | 签名 | 说明 |
|------|------|------|
| `SelectFile` | `() → ([]string, error)` | 系统原生多选文件对话框 |
| `OpenSaveDirectory` | `() → error` | `explorer` / `open` / `xdg-open` |
| `GetSaveDirectory` | `() → string` | 返回当前保存目录 |
| `SelectSaveDirectory` | `() → (string, error)` | 系统文件夹选择器 |
| `SetSaveDirectory` | `(newPath string) → error` | 手动设置 + 自动持久化 |

#### 网络/工具

| 方法 | 签名 | 说明 |
|------|------|------|
| `GetLocalIP` | `() → (string, error)` | UDP 伪连接 + 遍历网卡降级 |

### 9.4 内部函数（不暴露给前端）

| 函数 | 位置 | 说明 |
|------|------|------|
| `getConfigFilePath()` | `app.go` | 拼接配置文件路径，首次运行自动创建目录 |
| `NewApp()` | `app.go` | App 构造器，加载配置 |
| `loadConfig()` | `app.go` | 读 config.json，失败静默忽略 |
| `saveConfig()` | `app.go` | 写 config.json |
| `startUp(ctx)` | `app.go` | Wails 生命周期钩子，注入 ctx + 创建保存目录 |
| `zipFileToArchive(zw, srcPath, zipPath)` | `server_send.go` | 单文件 → ZIP 条目的辅助函数 |
| `getUploadHTML()` | `html_template.go` | 返回内嵌 H5 页面完整字符串 |

---

## 10. 前端架构

### 10.1 组件结构（`frontend/src/App.vue`）

```
App.vue (351 行)
├── Tab Header
│   ├── 📤 发送文件 button → switchMode('send')
│   └── 📥 接收文件 button → switchMode('receive')
│
├── 发送模式 (v-if="activeMode === 'send'")
│   ├── drop-zone → triggerFileSelect()
│   └── qr-container（选择文件后显示）
│       ├── qrcode-vue :value="shareUrl" :size="160"
│       ├── p.url-text → shareUrl
│       └── copyLink()
│
├── 接收模式 (v-else)
│   ├── start-receive-box → startReceive()
│   ├── qr-container（开启服务后显示）
│   │   ├── qrcode-vue :value="receiveUrl"
│   │   ├── btn-group → copyLink() / openFolder()
│   └── settings-box
│       ├── path-display（readonly input）
│       └── changeDirectory() → SelectSaveDirectory()
│
└── status-bar（底部浮动提示，error 类变红）
```

### 10.2 响应式状态

| ref | 类型 | 说明 |
|-----|------|------|
| `activeMode` | `'send' \| 'receive'` | 当前 Tab |
| `selectedFilePaths` | `string[]` | 发送模式下选中的文件绝对路径 |
| `selectedFileText` | `string` | 展示用（单文件显示文件名，多文件显示计数） |
| `shareUrl` | `string` | 发送模式生成的下载 URL |
| `receiveUrl` | `string` | 接收模式生成的上传页 URL |
| `savePath` | `string` | 当前保存目录（挂载时从后端读取） |
| `statusMsg` | `string` | 状态提示文本 |
| `isError` | `boolean` | 状态提示是否为错误样式 |
| `defaultPort` | `8080` | 硬编码端口 |

### 10.3 Go ↔ JS 通信

Wails 自动在 `window.go.main.App` 上暴露绑定方法，前端通过 `frontend/wailsjs/go/main/App.js` 调用：

```javascript
// 生成的代理代码示例（勿手动修改）
export function StartSendMode(arg1, arg2) {
  return window['go']['main']['App']['StartSendMode'](arg1, arg2);
}
```

**绑定规则**：
- 只有**导出方法**（首字母大写）才会被绑定
- 参数和返回值自动 JSON 序列化
- 错误通过 Go `error` 类型自动映射为 JS `reject`

### 10.4 onMounted 初始化

```javascript
onMounted(async () => {
  savePath.value = await GetSaveDirectory();  // 同步后端配置
});
```

没有"恢复上次模式"的逻辑 — 每次启动默认 Tab 都是发送模式。

---

## 11. 构建与部署

### 11.1 开发环境准备

```bash
# 安装 Go 1.25+
# 安装 Node.js 18+
# 安装 Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### 11.2 开发模式

```bash
cd LanShare
wails dev
```

效果：同时启动 Go 后端热重载 + Vite 前端 dev server，窗口会自动弹出。

### 11.3 生产构建

```bash
# 仅构建前端（会自动触发 Go embed 更新）
cd frontend && npm run build && cd ..

# 全量构建
wails build
```

产物：
```
build/bin/LanShare.exe     # 单文件可执行程序（包含前端资源）
```

### 11.4 Windows 安装包

`build/windows/installer/project.nsi` + `wails_tools.nsh` 定义了 NSIS 安装脚本，`wails build` 时会自动调用生成 `.exe` 安装包。

---

## 12. 并发与安全设计

### 12.1 并发保护

只有一个共享资源需要保护：`App.server`。

```go
serverLock sync.Mutex
```

保护点：
- `StopServer()` — 加锁后检查 `server != nil`，Shutdown 后置 nil
- `StartSendMode()` / `StartReceiveMode()` — 先调 StopServer（内部已加锁），再加自己的锁启动新服务

**竞态风险场景**：用户快速切换发送→接收→发送。没有锁保护的话，可能出现两个 goroutine 同时对同一个 `*http.Server` 操作，或者新旧 URL 错乱显示。

### 12.2 HTTP 安全

| 项目 | 当前状态 | 风险等级 |
|------|----------|----------|
| 传输层 | 明文 HTTP | ⚠️ 局域网内 OK，公网不行 |
| 认证 | 无 | ⚠️ 同网段任何人可访问 |
| CORS | 无限制 | ⚠️ 任意网页可发起请求 |
| CSRF | 无 token | ⚠️ 浏览器侧无法主动攻击（需要同一 IP） |
| 文件路径遍历 | `filepath.Join` + 直接拼接 | ⚠️ 低风险（接收方自己控制文件名） |

**设计假设**：工具定位为"同信任局域网"使用场景，所以安全优先级最低。如果要开放到公网，至少需要加 Token 认证 + HTTPS。

---

## 13. 当前已实现 vs 未实现

### ✅ 已实现

| 功能 | 位置 |
|------|------|
| 单文件直传（含断点续传） | `server_send.go` + `http.ServeContent` |
| 多文件/文件夹 ZIP 流式打包 | `server_send.go` + `archive/zip` |
| 内嵌 H5 上传页（拖拽 + 多文件） | `html_template.go` |
| 配置持久化（保存目录） | `app.go` + `config.json` |
| 跨平台 IP 获取 | `network.go` + UDP 伪连接 + 网卡遍历 |
| 跨平台打开文件夹 | `network.go` + `explorer/open/xdg-open` |
| 中文文件名支持 | `Content-Disposition` RFC 5987 |
| 并发安全 | `sync.Mutex` 保护 server |
| 防缓存 | `Cache-Control` + URL 时间戳参数 |

### ❌ 未实现（可作为迭代方向）

| 功能 | 现状 | 改造难度 |
|------|------|----------|
| 传输进度推送 | 前端只有"正在上传..."，无百分比 | 中（需加 SSE/WebSocket） |
| 断点续传（ZIP 模式） | 多文件 ZIP 不支持 Range | 高（需要自定义打包 + 分片） |
| 文件列表/历史记录 | 界面上看不到已接收了哪些文件 | 低（ReadDir + 列表渲染） |
| 传输取消 | 关闭窗口即停，无主动取消按钮 | 低（加 StopServer 调用即可） |
| 端口可配置 | 硬编码 8080 | 低（加配置项 + UI 输入框） |
| 发送方主动暂停/继续 | 无 | 中（需要 reader 包装 + 信号控制） |
| HTTPS / 认证 | 无 | 中（自签名证书 + 简单 Token） |
| 自动选择空闲端口 | 固定 8080，冲突直接报错 | 低（port 0 让 OS 分配） |
| 文件大小预校验 | 无（磁盘满才会报错） | 低（Statfs 检查剩余空间） |
| 暗黑模式 | 无 | 低（CSS 变量） |

---

## 14. 已知问题与边界情况

| # | 问题 | 触发条件 | 后果 | 建议 |
|---|------|----------|------|------|
| 1 | **中文文件夹名 ZIP 内路径分隔符** | Windows 下文件夹名含中文 | `filepath.ToSlash` 已处理，应该没问题 | 实测验证 |
| 2 | **大文件夹打包耗 CPU** | 选了一个含 10000 个小文件的目录 | CPU 飙升，浏览器端长时间无响应 | 可以加异步任务 + 进度推送 |
| 3 | **同文件名覆盖** | 手机上传了 `IMG_001.jpg`，之前电脑已有同名文件 | 直接覆盖，无冲突提示 | 加自动重命名（`IMG_001_1.jpg`） |
| 4 | **服务端口被占用** | 8080 已被其他程序占用 | `ListenAndServe` 报错，前端显示"启动失败" | 自动扫描下一个可用端口 |
| 5 | **上传文件名带路径** | 某些浏览器上传时 `fileHeader.Filename` 含 `\` | 直接拼接到 saveDirectory，可能产生子目录 | 上传前 `filepath.Base()` 清洗 |
| 6 | **防火墙首次弹窗** | Windows Defender 防火墙 | 用户可能不小心点拒绝 | 应用启动时主动申请规则 |

---

## 15. License

MIT © LanShare Contributors