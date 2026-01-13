package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"trae-proxy-go/internal/config"
	"trae-proxy-go/internal/logger"
	"trae-proxy-go/internal/proxy"
)

func main() {
	// 命令行参数
	var (
		configPath = flag.String("config", "config.yaml", "配置文件路径")
		certFile   = flag.String("cert", "", "证书文件路径")
		keyFile    = flag.String("key", "", "私钥文件路径")
		debug      = flag.Bool("debug", false, "启用调试模式")
	)
	flag.Parse()

	// 创建日志记录器
	log := logger.NewLogger(*debug)

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Error("加载配置失败: %v", err)
		os.Exit(1)
	}

	// 如果启用了调试模式，更新配置
	if *debug {
		cfg.Server.Debug = true
	}

	// 确定证书文件路径
	if *certFile == "" {
		*certFile = filepath.Join("ca", fmt.Sprintf("%s.crt", cfg.Domain))
	}
	if *keyFile == "" {
		*keyFile = filepath.Join("ca", fmt.Sprintf("%s.key", cfg.Domain))
	}

	// 检查证书文件
	if _, err := os.Stat(*certFile); os.IsNotExist(err) {
		log.Error("证书文件不存在: %s", *certFile)
		log.Info("请先运行证书生成工具生成证书")
		os.Exit(1)
	}
	if _, err := os.Stat(*keyFile); os.IsNotExist(err) {
		log.Error("私钥文件不存在: %s", *keyFile)
		log.Info("请先运行证书生成工具生成证书")
		os.Exit(1)
	}

	// 创建服务器
	srv, err := proxy.NewServer(cfg, log, *certFile, *keyFile)
	if err != nil {
		log.Error("创建服务器失败: %v", err)
		os.Exit(1)
	}

	// 启动服务器
	if err := srv.Start(); err != nil {
		log.Error("服务器启动失败: %v", err)
		os.Exit(1)
	}
}

