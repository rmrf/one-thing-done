package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"

	"github.com/BurntSushi/toml"
)

// Config 代表配置文件结构
type Config struct {
	Server   ServerConfig   `toml:"server"`
	Security SecurityConfig `toml:"security"`
	Commands []Command      `toml:"commands"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	APIKey string `toml:"apikey"`
}

// Command 命令配置
type Command struct {
	Slug string   `toml:"slug"`
	Desc string   `toml:"desc"`
	Cmd  string   `toml:"cmd"`
	Args []string `toml:"args"`
}

var config Config

func main() {
	// 解析命令行参数
	var configPath string
	flag.StringVar(&configPath, "config", "config.toml", "Path to config file")
	flag.Parse()

	// 读取配置文件
	if _, err := toml.DecodeFile(configPath, &config); err != nil {
		log.Fatalf("无法读取配置文件 %s: %v", configPath, err)
	}

	// 验证配置
	if config.Server.Port == 0 {
		config.Server.Port = 9090
	}
	if config.Server.Host == "" {
		config.Server.Host = "0.0.0.0"
	}

	// 创建命令映射
	commandMap := make(map[string]Command)
	log.Println("已加载的命令列表:")
	for _, cmd := range config.Commands {
		commandMap[cmd.Slug] = cmd
		if cmd.Desc != "" {
			log.Printf("  - %s: %s", cmd.Slug, cmd.Desc)
		} else {
			log.Printf("  - %s", cmd.Slug)
		}
	}

	// 设置路由
	http.HandleFunc("/one-thing-done/", func(w http.ResponseWriter, r *http.Request) {
		handleCommand(w, r, commandMap)
	})

	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	log.Printf("服务器启动在 http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func handleCommand(w http.ResponseWriter, r *http.Request, commandMap map[string]Command) {
	// 只接受 GET 请求
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 验证 API Key
	if config.Security.APIKey != "" {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			apiKey = r.URL.Query().Get("apikey")
		}
		if apiKey != config.Security.APIKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// 提取 slug
	path := strings.TrimPrefix(r.URL.Path, "/one-thing-done/")
	slug := strings.TrimSpace(path)

	if slug == "" {
		http.Error(w, "Missing command slug", http.StatusBadRequest)
		return
	}

	// 查找命令
	cmd, exists := commandMap[slug]
	if !exists {
		http.Error(w, fmt.Sprintf("Command '%s' not found", slug), http.StatusNotFound)
		return
	}

	// 执行命令
	log.Printf("执行命令: %s %v", cmd.Cmd, cmd.Args)

	execCmd := exec.Command(cmd.Cmd, cmd.Args...)
	output, err := execCmd.CombinedOutput()

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %v\n\nOutput:\n%s", err, string(output))
		return
	}

	w.WriteHeader(http.StatusOK)
	if len(output) == 0 {
		fmt.Fprintln(w, "Command executed successfully (no output)")
	} else {
		w.Write(output)
	}
}
