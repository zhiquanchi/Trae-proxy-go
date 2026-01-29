# Trae-Proxy Go

这是 Trae-Proxy 的 Golang 实现版本，一个高性能的 API 代理工具，专门用于拦截和重定向 OpenAI API 请求到自定义后端服务。

## 特性

- **智能代理**: 拦截 OpenAI API 请求并转发到自定义后端
- **多后端支持**: 配置多个 API 后端，支持动态切换
- **模型映射**: 自定义模型 ID 映射，无缝替换目标模型
- **流式响应**: 支持流式和非流式响应模式切换
- **SSL 证书**: 自动生成和管理自签名证书
- **自动配置**: 证书生成后自动安装CA证书和配置hosts文件
- **TUI 界面**: 友好的终端用户界面，方便配置管理
- **高性能**: 基于 Go 实现，性能优异

## 使用

### 系统要求

- Go 1.21 或更高版本
- OpenSSL（用于证书生成）

### 安装

```bash
# 克隆仓库
git clone <repository-url>
cd trae-proxy-go

# 安装依赖
go mod download

# 构建
go build -o trae-proxy ./cmd/proxy
go build -o trae-proxy-cli ./cmd/cli
```

### 配置

编辑 `config.yaml` 文件：

```yaml
domain: api.openai.com

apis:
  - name: "deepseek-r1"
    endpoint: "https://api.deepseek.com"
    custom_model_id: "deepseek-reasoner"
    target_model_id: "deepseek-reasoner"
    stream_mode: null
    active: true

server:
  port: 443
  debug: true
```

### 使用 TUI 界面（推荐）

直接运行 CLI 工具（无参数）将启动 TUI 界面：

```bash
./trae-proxy-cli
```

TUI 界面快捷键：
- `a` - 添加新的 API 配置
- `e` - 编辑选中的 API 配置
- `d` - 删除选中的 API 配置
- `空格` - 激活/停用选中的 API 配置
- `D` - 设置代理域名
- `C` - 生成 SSL 证书
- `q` - 退出
- `↑↓` - 上下选择

### 使用 CLI 命令

您也可以使用传统的 CLI 命令：

#### 生成证书

```bash
# 使用 CLI 工具生成证书
./trae-proxy-cli cert

# 或指定域名
./trae-proxy-cli cert --domain api.openai.com

# 生成证书并自动配置系统（推荐，需要管理员/root权限）
./trae-proxy-cli cert --auto-config

# 仅安装CA证书到系统信任存储
./trae-proxy-cli cert --install-ca

# 仅更新hosts文件
./trae-proxy-cli cert --update-hosts
```

**自动配置功能：**

使用 `--auto-config` 参数可以在证书生成后自动完成以下配置：
1. 安装CA证书到系统信任存储
2. 更新hosts文件，将域名指向 127.0.0.1

这样您无需手动配置，可以立即使用代理服务。

> 注意：自动配置需要管理员权限（Windows）或root权限（macOS/Linux）。
> 在TUI界面中生成证书后，也会提示是否进行自动配置。

### 启动服务器

```bash
# 使用编译后的二进制文件
./trae-proxy

# 或使用 go run
go run cmd/proxy/main.go

# 启用调试模式
./trae-proxy --debug
```

### CLI 工具使用

#### 列出配置

```bash
./trae-proxy-cli list
```

#### 添加 API 配置

```bash
./trae-proxy-cli add \
  --name "my-api" \
  --endpoint "https://api.example.com" \
  --custom-model "my-model" \
  --target-model "target-model" \
  --stream-mode none \
  --active
```

#### 更新 API 配置

```bash
./trae-proxy-cli update \
  --index 0 \
  --name "updated-name" \
  --endpoint "https://new-api.example.com"
```

#### 激活 API 配置

```bash
./trae-proxy-cli activate --index 0
```

#### 删除 API 配置

```bash
./trae-proxy-cli remove --index 0
```

#### 更新域名

```bash
./trae-proxy-cli domain --name api.openai.com
```

### 客户端配置

#### 与其他代理共存 / 冲突排查

如果你的机器上同时开着其他代理（如 Clash / Surge / 系统代理 / TUN 模式等），可能会出现两类常见冲突：

1. **端口冲突**：Trae-Proxy 默认监听 `443`（见 `config.yaml` 的 `server.port`）。如果已有程序占用 443，Trae-Proxy 将无法启动。
   - 解决：关闭占用 443 的程序，或修改 `server.port` 为其他端口（并确保客户端能配置对应的端口 / base_url），必要时使用端口转发。
2. **链路冲突（最常见）**：如果客户端走“系统代理/全局代理”，请求可能先被其他代理接管（例如 CONNECT `api.openai.com:443`），此时本机 `hosts` 指向 `127.0.0.1` 可能不会生效，流量到不了 Trae-Proxy。
   - 解决：在代理软件中把 `api.openai.com` 设置为 **直连/绕过代理**（bypass / DIRECT），并确保 `127.0.0.1`、`localhost` 不被代理。

##### 一键诊断

使用 CLI 的 `doctor` 命令检测环境/系统代理与端口占用，并生成建议设置：

```bash
./trae-proxy-cli doctor

# 可选：覆盖配置文件、域名、端口
./trae-proxy-cli doctor --config config.yaml --domain api.openai.com --port 443
```

> 说明：Trae-Proxy 转发到后端时会使用 Go 的默认代理规则，可能受进程环境变量 `HTTP_PROXY/HTTPS_PROXY/ALL_PROXY/NO_PROXY` 影响；`doctor` 会给出推荐的 `NO_PROXY` 值以避免请求被其他代理接管。

#### 1. 获取服务器自签证书

从服务器复制 CA 证书到本地：

```bash
scp user@your-server-ip:/path/to/trae-proxy-go/ca/ca.crt .
```

#### 2. 安装 CA 证书

##### Windows 系统

1. 双击 `ca.crt` 文件
2. 选择"安装证书"
3. 选择"本地计算机"
4. 选择"将所有证书放入下列存储" → "浏览" → "受信任的根证书颁发机构"
5. 完成安装

##### macOS 系统

1. 双击 `ca.crt` 文件，系统会打开"钥匙串访问"
2. 将证书添加到"系统"钥匙串
3. 双击导入的证书，展开"信任"部分
4. 将"使用此证书时"设置为"始终信任"
5. 关闭窗口并输入管理员密码确认

#### 3. 修改 hosts 文件

##### Windows 系统

编辑 `C:\Windows\System32\drivers\etc\hosts`，添加：

```
your-server-ip api.openai.com
```

> 如果 Trae-Proxy 就运行在本机，将 `your-server-ip` 替换为 `127.0.0.1`（或 `::1`）。

##### macOS 系统

编辑 `/etc/hosts`，添加：

```
your-server-ip api.openai.com
```

> 如果 Trae-Proxy 就运行在本机，将 `your-server-ip` 替换为 `127.0.0.1`（或 `::1`）。

#### 4. 测试连接

```bash
curl https://api.openai.com/v1/models
```

如果配置正确，您应该能看到代理服务器返回的模型列表。

## 项目结构与开发

### 项目结构

```
trae-proxy-go/
├── cmd/
│   ├── proxy/          # 主代理服务器程序
│   └── cli/            # 命令行管理工具
├── internal/
│   ├── config/         # 配置管理
│   ├── proxy/          # 代理核心逻辑
│   ├── cert/           # 证书管理
│   └── logger/         # 日志
├── pkg/
│   └── models/         # 数据模型
├── config.yaml         # 配置文件
├── ca/                 # 证书目录
└── README.md
```

### 实现原理

```
 +------------------+    +--------------+    +------------------+
 |                  |    |              |    |                  |
 |  DeepSeek API    +--->+              +--->+  Trae IDE        |
 |                  |    |              |    |                  |
 |  Moonshot API    +--->+              +--->+  VSCode          |
 |                  |    |  Trae-Proxy  |    |                  |
 |  Aliyun API      +--->+     Go       +--->+  JetBrains       |
 |                  |    |              |    |                  |
 |  Self-hosted LLM +--->+              +--->+  OpenAI Clients  |
 |                  |    |              |    |                  |
 +------------------+    +--------------+    +------------------+
   Backend Services       Proxy Server        Client Apps
```

### 与 Python 版本的差异

1. **性能**: Go 版本性能更好，资源占用更少
2. **部署**: 编译为单一二进制文件，无需 Python 运行时
3. **依赖**: 依赖更少，只需要 OpenSSL（用于证书生成）
4. **CLI 工具**: CLI 工具功能基本相同，但实现方式不同

### 开发

```bash
# 运行测试（如果有）
go test ./...

# 代码格式化
go fmt ./...

# 代码检查
go vet ./...
```

### GitHub Actions（CI / Release）

仓库内置 GitHub Actions 工作流：每次 push 到 `main`（以及手动触发 `workflow_dispatch`）都会自动运行 `go test ./...`，并构建 Windows/macOS/Linux 的可执行文件，最后把产物上传到 GitHub Releases 的 `nightly` 预发布版本。

触发方式参考文档：
https://docs.github.com/zh/actions/how-tos/write-workflows/choose-when-workflows-run/trigger-a-workflow

## 许可证

本项目采用 MIT 许可证。

## 贡献

欢迎提交 Issue 和 Pull Request！

