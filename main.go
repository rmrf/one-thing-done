package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

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

var (
	config      Config
	configMutex sync.RWMutex
	configPath  string
)

func main() {
	// 解析命令行参数
	flag.StringVar(&configPath, "config", "config.toml", "Path to config file")
	flag.Parse()

	// 初始加载配置
	if err := loadConfig(); err != nil {
		log.Fatalf("无法读取配置文件 %s: %v", configPath, err)
	}

	// 打印初始命令列表
	printCommandList()

	// 启动配置文件热加载监听
	go watchConfig()

	// 设置路由
	http.HandleFunc("/one-thing-done/", handleCommand)

	addr := fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
	log.Printf("服务器启动在 http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func loadConfig() error {
	var newConfig Config
	if _, err := toml.DecodeFile(configPath, &newConfig); err != nil {
		return err
	}

	// 验证配置
	if newConfig.Server.Port == 0 {
		newConfig.Server.Port = 9090
	}
	if newConfig.Server.Host == "" {
		newConfig.Server.Host = "0.0.0.0"
	}

	configMutex.Lock()
	config = newConfig
	configMutex.Unlock()

	return nil
}

func watchConfig() {
	var lastModTime time.Time

	// 获取初始修改时间
	if info, err := os.Stat(configPath); err == nil {
		lastModTime = info.ModTime()
	}

	for {
		time.Sleep(2 * time.Second)

		info, err := os.Stat(configPath)
		if err != nil {
			log.Printf("无法读取配置文件: %v", err)
			continue
		}

		if info.ModTime().After(lastModTime) {
			lastModTime = info.ModTime()
			log.Println("检测到配置文件变更，正在重新加载...")

			if err := loadConfig(); err != nil {
				log.Printf("重新加载配置文件失败: %v", err)
				continue
			}

			printCommandList()
			log.Println("配置文件已重新加载")
		}
	}
}

func printCommandList() {
	configMutex.RLock()
	defer configMutex.RUnlock()

	baseURL := fmt.Sprintf("http://%s:%d/one-thing-done/", config.Server.Host, config.Server.Port)
	log.Println("已加载的命令列表:")
	for _, cmd := range config.Commands {
		url := baseURL + cmd.Slug
		if cmd.Desc != "" {
			log.Printf("  - %s: %s", cmd.Desc, url)
		} else {
			log.Printf("  - %s", url)
		}
	}
}

func getCommandMap() map[string]Command {
	configMutex.RLock()
	defer configMutex.RUnlock()

	commandMap := make(map[string]Command)
	for _, cmd := range config.Commands {
		commandMap[cmd.Slug] = cmd
	}
	return commandMap
}

func handleCommand(w http.ResponseWriter, r *http.Request) {
	// 只接受 GET 请求
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	configMutex.RLock()
	apiKey := config.Security.APIKey
	configMutex.RUnlock()

	// 验证 API Key
	if apiKey != "" {
		requestApiKey := r.Header.Get("X-API-Key")
		if requestApiKey == "" {
			requestApiKey = r.URL.Query().Get("apikey")
		}
		if requestApiKey != apiKey {
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
	commandMap := getCommandMap()
	cmd, exists := commandMap[slug]
	if !exists {
		http.Error(w, fmt.Sprintf("Command '%s' not found", slug), http.StatusNotFound)
		return
	}

	// 执行命令
	log.Printf("执行命令: %s %v", cmd.Cmd, cmd.Args)

	execCmd := exec.Command(cmd.Cmd, cmd.Args...)
	// 设置环境变量，避免 TTY 相关问题
	execCmd.Env = append(os.Environ(), "TERM=dumb")
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
