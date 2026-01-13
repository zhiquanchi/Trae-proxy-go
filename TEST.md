# 测试检查清单

由于系统上未安装Go，这里提供一个测试检查清单，供在有Go环境时使用。

## 编译测试

```bash
# 1. 下载依赖
go mod download

# 2. 编译代理服务器
go build -o trae-proxy ./cmd/proxy

# 3. 编译CLI工具
go build -o trae-proxy-cli ./cmd/cli

# 4. 检查编译是否成功
./trae-proxy --help
./trae-proxy-cli
```

## 代码检查

```bash
# 格式化代码
go fmt ./...

# 静态分析
go vet ./...

# 检查未使用的导入
goimports -w .
```

## 功能测试

### 1. 配置管理测试

```bash
# 列出配置
./trae-proxy-cli list

# 添加API配置
./trae-proxy-cli add \
  --name "test-api" \
  --endpoint "https://api.openai.com" \
  --custom-model "test-model" \
  --target-model "gpt-4" \
  --stream-mode none

# 激活API配置
./trae-proxy-cli activate --index 0

# 更新API配置
./trae-proxy-cli update --index 0 --name "updated-api"

# 删除API配置（如果有多于1个）
# ./trae-proxy-cli remove --index 1
```

### 2. 证书生成测试

```bash
# 生成证书
./trae-proxy-cli cert

# 检查证书文件是否存在
ls -la ca/
```

### 3. 服务器启动测试

```bash
# 启动服务器（需要先生成证书）
./trae-proxy --debug

# 在另一个终端测试
curl -k https://localhost/v1/models
curl -k https://localhost/
```

## 集成测试

### 测试代理功能

1. 确保配置了有效的后端API
2. 启动服务器
3. 使用curl测试各个端点：

```bash
# 测试根路径
curl -k https://localhost/

# 测试v1路径
curl -k https://localhost/v1

# 测试模型列表
curl -k https://localhost/v1/models

# 测试聊天完成（需要有效的API密钥）
curl -k -X POST https://localhost/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "test-model",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

## 已知问题和注意事项

1. **TLS配置**: server.go中使用`ListenAndServeTLS("", "")`是合法的，因为证书已通过TLSConfig设置
2. **流式响应**: 流式响应直接转发，不进行模型ID替换（与Python版本一致）
3. **Content-Type检查**: handler.go中检查`Content-Type`必须是`application/json`，这是严格的

## 性能测试

```bash
# 使用ab或wrk进行压力测试
ab -n 1000 -c 10 -k https://localhost/v1/models
```

## Docker测试

```bash
# 构建Docker镜像
docker build -t trae-proxy-go .

# 运行容器
docker run -d \
  -p 443:443 \
  -v $(pwd)/ca:/app/ca \
  -v $(pwd)/config.yaml:/app/config.yaml \
  --name trae-proxy-test \
  trae-proxy-go

# 查看日志
docker logs trae-proxy-test

# 测试
curl -k https://localhost/v1/models

# 清理
docker stop trae-proxy-test
docker rm trae-proxy-test
```

