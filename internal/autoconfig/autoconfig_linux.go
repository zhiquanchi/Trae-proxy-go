package autoconfig

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// installCACert 在Linux上安装CA证书到系统信任存储
// certPath: CA证书文件的完整路径
// 支持Debian/Ubuntu (update-ca-certificates) 和 Red Hat/CentOS (update-ca-trust)
// 注意：需要sudo权限
func installCACert(certPath string) error {
	// 不同Linux发行版可能有不同的证书目录
	// 这里使用最常见的方法：Debian/Ubuntu系列
	targetDir := "/usr/local/share/ca-certificates"
	targetPath := filepath.Join(targetDir, "trae-proxy-ca.crt")

	// 检查目标目录是否存在
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		// 尝试Red Hat/CentOS系列的路径
		targetDir = "/etc/pki/ca-trust/source/anchors"
		targetPath = filepath.Join(targetDir, "trae-proxy-ca.crt")

		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			return fmt.Errorf("找不到系统CA证书目录，请手动安装")
		}

		// Red Hat/CentOS系列
		return installCACertRedHat(certPath, targetPath)
	}

	// Debian/Ubuntu系列
	return installCACertDebian(certPath, targetPath)
}

// installCACertDebian 在Debian/Ubuntu上安装CA证书
func installCACertDebian(certPath, targetPath string) error {
	// 复制证书文件
	cmd := exec.Command("sudo", "cp", certPath, targetPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("复制证书失败: %w, 输出: %s", err, string(output))
	}

	// 更新CA证书
	cmd = exec.Command("sudo", "update-ca-certificates")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("更新CA证书失败: %w, 输出: %s", err, string(output))
	}

	return nil
}

// installCACertRedHat 在Red Hat/CentOS上安装CA证书
func installCACertRedHat(certPath, targetPath string) error {
	// 复制证书文件
	cmd := exec.Command("sudo", "cp", certPath, targetPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("复制证书失败: %w, 输出: %s", err, string(output))
	}

	// 更新CA证书
	cmd = exec.Command("sudo", "update-ca-trust")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("更新CA证书失败: %w, 输出: %s", err, string(output))
	}

	return nil
}
