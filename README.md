# One Thing Done

一个简单的 Go 服务，通过 HTTP 请求在目标机器上执行预配置的命令。

## 功能

- 通过 `config.toml` 配置可执行的命令
- 支持 API Key 验证
- 每个命令通过唯一的 slug 标识
- 返回命令执行结果
- 配置文件热加载（无需重启服务）
- 启动时显示所有可用的命令 URL

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
desc = "重启 OpenClash 服务"
cmd = "systemctl"
args = ["restart", "openclaw.service"]
```

### 配置文件热加载

服务启动后会自动监听配置文件的变化，每 2 秒检查一次。当配置文件被修改时，服务会自动重新加载配置并更新命令列表，无需重启服务。

```
2026/03/22 10:25:37 检测到配置文件变更，正在重新加载...
2026/03/22 10:25:37 已加载的命令列表:
2026/03/22 10:25:37   - 查看当前时间: http://0.0.0.0:9090/one-thing-done/date
2026/03/22 10:25:37 配置文件已重新加载
```

## 使用

### 命令行参数

```bash
./one-thing-done --config /path/to/config.toml
```

- `--config`: 配置文件路径（默认: `config.toml`）

### 启动服务

```bash
# 使用默认配置文件（当前目录的 config.toml）
./one-thing-done

# 指定配置文件
./one-thing-done --config /usr/local/etc/one-thing-done-config.toml
```

### Systemd 安装

```bash
# 复制二进制文件
sudo cp one-thing-done /usr/local/bin/
sudo chmod +x /usr/local/bin/one-thing-done

# 复制配置文件
sudo cp config.toml /usr/local/etc/one-thing-done-config.toml
sudo chmod 600 /usr/local/etc/one-thing-done-config.toml

# 复制 service 文件
sudo cp one-thing-done.service /etc/systemd/system/

# 启动服务
sudo systemctl daemon-reload
sudo systemctl enable one-thing-done
sudo systemctl start one-thing-done

# 查看状态
sudo systemctl status one-thing-done
```

### 发送请求
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

## 许可证

MIT License
