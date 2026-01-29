package autoconfig

import (
	"fmt"
	"os/exec"
)

// installCACert 在macOS上安装CA证书
func installCACert(certPath string) error {
	// 使用security命令将证书添加到系统钥匙串并设置为受信任
	cmd := exec.Command("sudo", "security", "add-trusted-cert", 
		"-d", "-r", "trustRoot", 
		"-k", "/Library/Keychains/System.keychain", 
		certPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("执行security命令失败: %w, 输出: %s", err, string(output))
	}
	
	return nil
}
