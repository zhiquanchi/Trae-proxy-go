package cert

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// GenerateCertificates 生成CA证书和服务器证书
func GenerateCertificates(domain string, caDir string) error {
	// 创建ca目录（如果不存在）
	if err := os.MkdirAll(caDir, 0755); err != nil {
		return fmt.Errorf("创建证书目录失败: %w", err)
	}

	// 检查OpenSSL是否可用
	if err := checkOpenSSL(); err != nil {
		return err
	}

	// 生成CA证书（如果不存在）
	caKeyPath := filepath.Join(caDir, "ca.key")
	caCertPath := filepath.Join(caDir, "ca.crt")
	if _, err := os.Stat(caKeyPath); os.IsNotExist(err) {
		if err := generateCA(caDir); err != nil {
			return fmt.Errorf("生成CA证书失败: %w", err)
		}
	}

	// 生成服务器证书
	if err := generateServerCert(domain, caDir, caKeyPath, caCertPath); err != nil {
		return fmt.Errorf("生成服务器证书失败: %w", err)
	}

	return nil
}

// checkOpenSSL 检查OpenSSL是否已安装
func checkOpenSSL() error {
	cmd := exec.Command("openssl", "version")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("未找到OpenSSL，请确保OpenSSL已安装并在PATH中")
	}
	return nil
}

// generateCA 生成CA证书和私钥
func generateCA(caDir string) error {
	caKeyPath := filepath.Join(caDir, "ca.key")
	caCertPath := filepath.Join(caDir, "ca.crt")

	// 生成CA私钥
	cmd := exec.Command("openssl", "genrsa", "-out", caKeyPath, "2048")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("生成CA私钥失败: %w", err)
	}

	// 生成CA证书
	subject := "/C=CN/ST=State/L=City/O=TraeProxy CA/OU=TraeProxy/CN=TraeProxy Root CA"
	cmd = exec.Command("openssl", "req", "-new", "-x509", "-days", "36500",
		"-key", caKeyPath, "-out", caCertPath, "-subj", subject)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("生成CA证书失败: %w", err)
	}

	return nil
}

// generateServerCert 生成服务器证书
func generateServerCert(domain, caDir, caKeyPath, caCertPath string) error {
	keyPath := filepath.Join(caDir, fmt.Sprintf("%s.key", domain))
	certPath := filepath.Join(caDir, fmt.Sprintf("%s.crt", domain))
	csrPath := filepath.Join(caDir, fmt.Sprintf("%s.csr", domain))
	cnfPath := filepath.Join(caDir, fmt.Sprintf("%s.cnf", domain))

	// 创建OpenSSL配置文件
	cnfContent := fmt.Sprintf(`[ req ]
default_bits        = 2048
default_md          = sha256
distinguished_name  = req_distinguished_name
req_extensions      = v3_req

[ req_distinguished_name ]

[ v3_req ]
basicConstraints       = CA:FALSE
keyUsage               = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage       = serverAuth
subjectAltName         = @alt_names

[ alt_names ]
DNS.1 = %s
`, domain)

	if err := os.WriteFile(cnfPath, []byte(cnfContent), 0644); err != nil {
		return fmt.Errorf("创建配置文件失败: %w", err)
	}

	// 生成服务器私钥
	cmd := exec.Command("openssl", "genrsa", "-out", keyPath, "2048")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("生成服务器私钥失败: %w", err)
	}

	// 生成CSR
	subject := fmt.Sprintf("/C=CN/ST=State/L=City/O=Organization/OU=Unit/CN=%s", domain)
	cmd = exec.Command("openssl", "req", "-new", "-key", keyPath,
		"-out", csrPath, "-config", cnfPath, "-subj", subject)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("生成CSR失败: %w", err)
	}

	// 使用CA签署证书
	cmd = exec.Command("openssl", "x509", "-req", "-days", "365",
		"-in", csrPath, "-CA", caCertPath, "-CAkey", caKeyPath,
		"-CAcreateserial", "-out", certPath, "-extensions", "v3_req",
		"-extfile", cnfPath)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("签署证书失败: %w", err)
	}

	// 清理CSR文件
	os.Remove(csrPath)

	return nil
}

