# One Thing Done

一个简单的 Go 服务，通过 HTTP 请求在目标机器上执行预配置的命令。

## 功能

- 通过 `config.toml` 配置可执行的命令
- 支持 API Key 验证
- 每个命令通过唯一的 slug 标识
- 返回命令执行结果

## 安装

```bash
go mod tidy
go build -o one-thing-done
```

## 配置

编辑 `config.toml`:

```toml
[server]
host = "0.0.0.0"
port = 9090

[security]
apikey = "your-secret-api-key"

[[commands]]
slug = "restart_openclaw"
cmd = "systemctl"
args = ["restart", "openclaw.service"]
```

## 使用

启动服务:
```bash
./one-thing-done
```

发送请求:
```bash
# 使用 Header 传递 API Key
curl -H "X-API-Key: your-secret-api-key" http://localhost:9090/one-thing-done/restart_openclaw

# 或使用 Query 参数
curl http://localhost:9090/one-thing-done/restart_openclaw?apikey=your-secret-api-key
```

## API

- **URL**: `/one-thing-done/{slug}`
- **Method**: `GET`
- **Auth**: `X-API-Key` Header 或 `apikey` Query 参数
- **Response**: 纯文本格式的命令输出
