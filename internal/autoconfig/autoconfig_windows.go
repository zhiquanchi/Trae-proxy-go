package autoconfig

import (
	"fmt"
	"os/exec"
)

// installCACert 在Windows上安装CA证书
func installCACert(certPath string) error {
	// 使用certutil命令安装证书到受信任的根证书颁发机构
	cmd := exec.Command("certutil", "-addstore", "-f", "Root", certPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("执行certutil失败: %w, 输出: %s", err, string(output))
	}
	return nil
}
