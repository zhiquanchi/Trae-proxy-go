package autoconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// AutoConfigure 在证书生成后自动配置系统
// domain: 要代理的域名
// caDir: CA证书目录
// installCA: 是否安装CA证书到系统信任存储
// updateHosts: 是否更新hosts文件
func AutoConfigure(domain, caDir string, installCA, updateHosts bool) error {
	var errors []error

	// 安装CA证书
	if installCA {
		if err := installCACertificate(caDir); err != nil {
			errors = append(errors, fmt.Errorf("安装CA证书失败: %w", err))
		}
	}

	// 更新hosts文件
	if updateHosts {
		if err := updateHostsFile(domain); err != nil {
			errors = append(errors, fmt.Errorf("更新hosts文件失败: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("自动配置遇到错误: %v", errors)
	}

	return nil
}

// installCACertificate 安装CA证书到系统信任存储
func installCACertificate(caDir string) error {
	caCertPath := filepath.Join(caDir, "ca.crt")
	
	// 检查证书文件是否存在
	if _, err := os.Stat(caCertPath); os.IsNotExist(err) {
		return fmt.Errorf("CA证书文件不存在: %s", caCertPath)
	}

	return installCACert(caCertPath)
}

// updateHostsFile 更新hosts文件以将域名指向localhost
func updateHostsFile(domain string) error {
	var hostsPath string
	switch runtime.GOOS {
	case "windows":
		hostsPath = filepath.Join(os.Getenv("SystemRoot"), "System32", "drivers", "etc", "hosts")
	default: // darwin, linux
		hostsPath = "/etc/hosts"
	}

	// 读取现有hosts文件
	content, err := os.ReadFile(hostsPath)
	if err != nil {
		return fmt.Errorf("读取hosts文件失败: %w", err)
	}

	// 检查是否已经存在该域名的配置
	contentStr := string(content)
	entry := fmt.Sprintf("127.0.0.1 %s", domain)
	
	if containsHostsEntry(contentStr, domain) {
		// 已存在，无需添加
		return nil
	}

	// 添加新条目
	newContent := contentStr
	if len(newContent) > 0 && newContent[len(newContent)-1] != '\n' {
		newContent += "\n"
	}
	newContent += fmt.Sprintf("# Added by Trae-Proxy\n%s\n", entry)

	// 写入hosts文件（需要管理员权限）
	if err := os.WriteFile(hostsPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("写入hosts文件失败（需要管理员/root权限）: %w", err)
	}

	return nil
}

// containsHostsEntry 检查hosts文件内容是否已包含该域名
func containsHostsEntry(content, domain string) bool {
	lines := splitLines(content)
	for _, line := range lines {
		// 跳过注释和空行
		trimmed := trimSpace(line)
		if len(trimmed) == 0 || trimmed[0] == '#' {
			continue
		}
		
		// 检查是否包含该域名
		if containsString(trimmed, domain) {
			return true
		}
	}
	return false
}

// splitLines 将字符串按换行符分割
func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// trimSpace 去除字符串前后的空白字符
func trimSpace(s string) string {
	start := 0
	end := len(s)
	
	// 去除前导空白
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\r') {
		start++
	}
	
	// 去除尾部空白
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}
	
	return s[start:end]
}

// containsString 检查字符串是否包含子串
func containsString(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// NeedsElevatedPrivileges 检查是否需要提升权限
func NeedsElevatedPrivileges() bool {
	// 在所有平台上，安装CA证书和修改hosts文件都需要管理员权限
	return true
}

// GetInstructions 获取手动配置说明（用于无法自动配置时）
func GetInstructions(domain, caDir string) string {
	caCertPath := filepath.Join(caDir, "ca.crt")
	
	var instructions string
	switch runtime.GOOS {
	case "windows":
		instructions = fmt.Sprintf(`Windows 手动配置说明：

1. 安装CA证书：
   - 右键点击 %s
   - 选择"安装证书"
   - 选择"本地计算机"
   - 选择"将所有证书放入下列存储" → "浏览" → "受信任的根证书颁发机构"
   - 完成安装

2. 修改hosts文件：
   - 以管理员身份打开记事本
   - 打开文件: C:\Windows\System32\drivers\etc\hosts
   - 添加以下行:
     127.0.0.1 %s

3. 重启浏览器或应用程序使更改生效
`, caCertPath, domain)
	case "darwin":
		instructions = fmt.Sprintf(`macOS 手动配置说明：

1. 安装CA证书：
   sudo security add-trusted-cert -d -r trustRoot -k /Library/Keychains/System.keychain %s

2. 修改hosts文件：
   sudo sh -c 'echo "127.0.0.1 %s" >> /etc/hosts'

3. 刷新DNS缓存：
   sudo dscacheutil -flushcache
   sudo killall -HUP mDNSResponder
`, caCertPath, domain)
	case "linux":
		instructions = fmt.Sprintf(`Linux 手动配置说明：

1. 安装CA证书：
   sudo cp %s /usr/local/share/ca-certificates/trae-proxy-ca.crt
   sudo update-ca-certificates

2. 修改hosts文件：
   sudo sh -c 'echo "127.0.0.1 %s" >> /etc/hosts'

注意：某些Linux发行版可能需要不同的命令
`, caCertPath, domain)
	default:
		instructions = fmt.Sprintf("不支持的操作系统: %s\n请手动配置CA证书和hosts文件", runtime.GOOS)
	}
	
	return instructions
}
