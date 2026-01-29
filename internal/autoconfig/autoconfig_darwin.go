package autoconfig

import (
	"fmt"
	"os/exec"
)

// installCACert 在macOS上安装CA证书到系统钥匙串
// certPath: CA证书文件的完整路径
// 使用macOS的security命令将证书添加到系统钥匙串并设置为受信任
// 注意：需要sudo权限
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
