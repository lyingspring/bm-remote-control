package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	stdruntime "runtime"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type App struct {
	ctx         context.Context
	httpServer  *http.Server
	commandChan chan string
	resultChan  chan string
}

type CommandRequest struct {
	Command string `json:"command"`
}

type CommandResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
	Error   string `json:"error,omitempty"`
}

func NewApp() *App {
	return &App{
		commandChan: make(chan string, 100),
		resultChan:  make(chan string, 100),
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.startHTTPServer()
}

func (a *App) shutdown(ctx context.Context) {
	if a.httpServer != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		a.httpServer.Shutdown(shutdownCtx)
	}
}

func (a *App) startHTTPServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/shutdown", a.handleShutdown)
	mux.HandleFunc("/api/restart", a.handleRestart)
	mux.HandleFunc("/api/sleep", a.handleSleep)
	mux.HandleFunc("/api/command", a.handleCommand)
	mux.HandleFunc("/api/status", a.handleStatus)
	mux.HandleFunc("/health", a.handleHealth)

	a.httpServer = &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		fmt.Println("HTTP Server starting on :8080")
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP Server error: %v\n", err)
		}
	}()
}

func (a *App) handleShutdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"success": true, "message": "Shutdown command sent"}`)

	fmt.Println("Shutdown command received via HTTP")
	go a.executeShutdown()
}

func (a *App) handleRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"success": true, "message": "Restart command sent"}`)

	fmt.Println("Restart command received via HTTP")
	go a.executeRestart()
}

func (a *App) handleSleep(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"success": true, "message": "Sleep command sent"}`)

	fmt.Println("Sleep command received via HTTP")
	go a.executeSleep()
}

func (a *App) handleCommand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	command := string(body)
	if command == "" {
		http.Error(w, "Command is required", http.StatusBadRequest)
		return
	}

	fmt.Printf("Command received via HTTP: %s\n", command)

	output, err := a.executeCommand(command)
	response := CommandResponse{
		Success: err == nil,
		Output:  output,
	}

	if err != nil {
		response.Error = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"success": %t, "output": %q, "error": %q}`,
		response.Success, response.Output, response.Error)
}

func (a *App) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status": "running", "os": "%s", "arch": "%s"}`, stdruntime.GOOS, stdruntime.GOARCH)
}

func (a *App) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

func (a *App) executeShutdown() error {
	var cmd *exec.Cmd

	switch stdruntime.GOOS {
	case "darwin":
		cmd = exec.Command("shutdown", "-h", "now")
	case "windows":
		cmd = exec.Command("shutdown", "/s", "/t", "0")
	case "linux":
		cmd = exec.Command("shutdown", "-h", "now")
	default:
		return fmt.Errorf("unsupported operating system: %s", stdruntime.GOOS)
	}

	return cmd.Run()
}

func (a *App) executeRestart() error {
	var cmd *exec.Cmd

	switch stdruntime.GOOS {
	case "darwin":
		cmd = exec.Command("shutdown", "-r", "now")
	case "windows":
		cmd = exec.Command("shutdown", "/r", "/t", "0")
	case "linux":
		cmd = exec.Command("shutdown", "-r", "now")
	default:
		return fmt.Errorf("unsupported operating system: %s", stdruntime.GOOS)
	}

	return cmd.Run()
}

func (a *App) executeSleep() error {
	var cmd *exec.Cmd

	switch stdruntime.GOOS {
	case "darwin":
		cmd = exec.Command("pmset", "sleepnow")
	case "windows":
		cmd = exec.Command("rundll32.exe", "powrprof.dll,SetSuspendState", "0,1,0")
	case "linux":
		cmd = exec.Command("systemctl", "suspend")
	default:
		return fmt.Errorf("unsupported operating system: %s", stdruntime.GOOS)
	}

	return cmd.Run()
}

func (a *App) executeCommand(command string) (string, error) {
	var cmd *exec.Cmd

	if stdruntime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		parts := strings.Fields(command)
		if len(parts) == 0 {
			return "", fmt.Errorf("empty command")
		}
		cmd = exec.Command(parts[0], parts[1:]...)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

func (a *App) ShutdownComputer() (string, error) {
	fmt.Println("Shutdown requested from UI")
	go a.executeShutdown()
	return "Shutdown command sent", nil
}

func (a *App) RestartComputer() (string, error) {
	fmt.Println("Restart requested from UI")
	go a.executeRestart()
	return "Restart command sent", nil
}

func (a *App) SleepComputer() (string, error) {
	fmt.Println("Sleep requested from UI")
	err := a.executeSleep()
	if err != nil {
		return "", err
	}
	return "Sleep command sent", nil
}

func (a *App) ExecuteCommand(command string) (string, error) {
	fmt.Printf("Command from UI: %s\n", command)
	return a.executeCommand(command)
}

func (a *App) GetSystemInfo() (map[string]interface{}, error) {
	info := make(map[string]interface{})
	info["os"] = stdruntime.GOOS
	info["arch"] = stdruntime.GOARCH
	info["hostname"], _ = os.Hostname()

	var memStats stdruntime.MemStats
	stdruntime.ReadMemStats(&memStats)
	info["memory_alloc"] = memStats.Alloc
	info["memory_sys"] = memStats.Sys

	return info, nil
}

func (a *App) GetServerPort() string {
	return ":8080"
}

func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, Remote Control System is ready!", name)
}

// createSSHClient 创建 SSH 客户端连接
func (a *App) createSSHClient(config SSHConfig) (*ssh.Client, error) {
	// 构建 SSH 配置
	sshConfig := &ssh.ClientConfig{
		User: config.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境应该使用 ssh.FixedHostKey
		Timeout:         10 * time.Second,
	}

	// 如果有密码，使用密码认证
	if config.Password != "" {
		sshConfig.Auth = []ssh.AuthMethod{
			ssh.Password(config.Password),
		}
	} else {
		// 否则尝试使用 SSH agent 或密钥文件
		// 尝试使用 SSH agent
		if sshAgent, err := getSSHAgentAuth(); err == nil && sshAgent != nil {
			sshConfig.Auth = []ssh.AuthMethod{sshAgent}
		} else {
			// 尝试使用默认的私钥文件
			if keyAuth, err := getPublicKeyAuth(); err == nil && keyAuth != nil {
				sshConfig.Auth = []ssh.AuthMethod{keyAuth}
			} else {
				return nil, fmt.Errorf("没有可用的认证方式：请配置密码或设置 SSH 密钥")
			}
		}
	}

	// 连接到 SSH 服务器
	address := fmt.Sprintf("%s:%s", config.Host, config.Port)
	client, err := ssh.Dial("tcp", address, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("SSH 连接失败: %v", err)
	}

	return client, nil
}

// getSSHAgentAuth 尝试从 SSH agent 获取认证
func getSSHAgentAuth() (ssh.AuthMethod, error) {
	// 这里简化处理，实际可以使用 golang.org/x/crypto/ssh/agent
	// 为了简化，我们直接返回错误，让系统回退到密钥文件
	return nil, fmt.Errorf("SSH agent not available")
}

// getPublicKeyAuth 从默认密钥文件获取公钥认证
func getPublicKeyAuth() (ssh.AuthMethod, error) {
	// 尝试常见的私钥文件位置
	keyFiles := []string{
		os.Getenv("HOME") + "/.ssh/id_ed25519",
		os.Getenv("HOME") + "/.ssh/id_rsa",
		os.Getenv("HOME") + "/.ssh/id_ecdsa",
	}

	for _, keyFile := range keyFiles {
		key, err := os.ReadFile(keyFile)
		if err != nil {
			continue
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			// 尝试解析带密码的密钥
			if _, ok := err.(*ssh.PassphraseMissingError); ok {
				continue // 跳过需要密码的密钥
			}
			continue
		}

		return ssh.PublicKeys(signer), nil
	}

	return nil, fmt.Errorf("no valid key found")
}

func (a *App) SaveSSHConfig(host, port, username, password string) (string, error) {
	config := SSHConfig{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化配置失败: %v", err)
	}

	err = os.WriteFile("settings.json", data, 0600)
	if err != nil {
		return "", fmt.Errorf("保存配置失败: %v", err)
	}

	fmt.Printf("SSH 配置已保存: %s@%s:%s\n", username, host, port)
	return "SSH 配置已成功保存", nil
}

func (a *App) LoadSSHConfig() (map[string]interface{}, error) {
	data, err := os.ReadFile("settings.json")
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{
				"host":     "",
				"port":     "22",
				"username": "",
				"password": "",
			}, nil
		}
		return nil, fmt.Errorf("读取配置失败: %v", err)
	}

	var config SSHConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置失败: %v", err)
	}

	return map[string]interface{}{
		"host":     config.Host,
		"port":     config.Port,
		"username": config.Username,
		"password": config.Password,
	}, nil
}

func (a *App) ExecuteSSHCommand(command string) (string, error) {
	configData, err := os.ReadFile("settings.json")
	if err != nil {
		return "", fmt.Errorf("未找到 SSH 配置，请先配置连接信息")
	}

	var config SSHConfig
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return "", fmt.Errorf("解析配置失败: %v", err)
	}

	if config.Host == "" || config.Username == "" {
		return "", fmt.Errorf("SSH 配置不完整，请检查配置")
	}

	// 使用纯 Go SSH 客户端连接
	client, err := a.createSSHClient(config)
	if err != nil {
		return "", err
	}
	defer client.Close()

	// 创建会话
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建 SSH 会话失败: %v", err)
	}
	defer session.Close()

	// 检查是否是 sudo 命令且配置了密码
	if strings.HasPrefix(command, "sudo ") && config.Password != "" {
		// 添加 -S 参数让 sudo 从 stdin 读取密码
		sudoCommand := strings.Replace(command, "sudo ", "sudo -S ", 1)

		// 使用 stdin 输入密码给 sudo
		stdin, err := session.StdinPipe()
		if err != nil {
			return "", fmt.Errorf("创建 stdin 管道失败: %v", err)
		}

		// 收集输出
		var outputBuf bytes.Buffer
		session.Stdout = &outputBuf
		session.Stderr = &outputBuf

		// 启动命令
		if err := session.Start(sudoCommand); err != nil {
			return "", fmt.Errorf("启动命令失败: %v", err)
		}

		// 输入密码（立即输入，不需要等待）
		go func() {
			defer stdin.Close()
			fmt.Fprintf(stdin, "%s\n", config.Password)
		}()

		// 等待命令完成
		err = session.Wait()

		output := outputBuf.String()
		if err != nil {
			// 如果 sudo 密码错误，尝试不使用密码的方式
			if strings.Contains(output, "incorrect password") || strings.Contains(output, "Sorry, try again") {
				// 返回到普通命令执行
				output2, err2 := session.CombinedOutput(command)
				if err2 == nil {
					return string(output2), nil
				}
			}
			return output, fmt.Errorf("命令执行失败: %v", err)
		}

		return output, nil
	}

	// 执行普通命令（非 sudo 或无密码）
	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("命令执行失败: %v", err)
	}

	return string(output), nil
}

func (a *App) TestSSHConnection() (string, error) {
	configData, err := os.ReadFile("settings.json")
	if err != nil {
		return "", fmt.Errorf("未找到 SSH 配置，请先配置连接信息")
	}

	var config SSHConfig
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return "", fmt.Errorf("解析配置失败: %v", err)
	}

	if config.Host == "" || config.Username == "" {
		return "", fmt.Errorf("SSH 配置不完整，请检查配置")
	}

	// 使用纯 Go SSH 客户端连接测试
	client, err := a.createSSHClient(config)
	if err != nil {
		return "", err
	}
	defer client.Close()

	// 创建会话并测试
	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建 SSH 会话失败: %v", err)
	}
	defer session.Close()

	// 执行简单测试命令
	output, err := session.CombinedOutput("echo 'Connection successful'")
	if err != nil {
		return "", fmt.Errorf("SSH 连接测试失败: %v\n输出: %s", err, string(output))
	}

	return "SSH 连接测试成功！", nil
}

func (a *App) GetRemoteSystemInfo() (map[string]interface{}, error) {
	configData, err := os.ReadFile("settings.json")
	if err != nil {
		return nil, fmt.Errorf("未找到 SSH 配置，请先配置连接信息")
	}

	var config SSHConfig
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置失败: %v", err)
	}

	if config.Host == "" || config.Username == "" {
		return nil, fmt.Errorf("SSH 配置不完整，请检查配置")
	}

	// 使用纯 Go SSH 客户端连接
	client, err := a.createSSHClient(config)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	info := make(map[string]interface{})

	// 获取主机名
	session, err := client.NewSession()
	if err == nil {
		if hostname, err := session.CombinedOutput("hostname"); err == nil {
			info["hostname"] = strings.TrimSpace(string(hostname))
		}
		session.Close()
	}

	// 获取操作系统信息
	session, err = client.NewSession()
	if err == nil {
		if osInfo, err := session.CombinedOutput("uname -s"); err == nil {
			info["os"] = strings.TrimSpace(string(osInfo))
		}
		session.Close()
	}

	// 获取架构信息
	session, err = client.NewSession()
	if err == nil {
		if arch, err := session.CombinedOutput("uname -m"); err == nil {
			info["arch"] = strings.TrimSpace(string(arch))
		}
		session.Close()
	}

	// 获取系统运行时间
	session, err = client.NewSession()
	if err == nil {
		if uptime, err := session.CombinedOutput("uptime -p 2>/dev/null || uptime"); err == nil {
			info["uptime"] = strings.TrimSpace(string(uptime))
		}
		session.Close()
	}

	return info, nil
}
