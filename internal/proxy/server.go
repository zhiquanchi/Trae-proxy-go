package proxy

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"trae-proxy-go/internal/logger"
	"trae-proxy-go/pkg/models"
)

// Server 代理服务器
type Server struct {
	config   *models.Config
	logger   *logger.Logger
	handler  *Handler
	tlsConfig *tls.Config
}

// NewServer 创建新的代理服务器
func NewServer(config *models.Config, logger *logger.Logger, certFile, keyFile string) (*Server, error) {
	handler := NewHandler(config, logger)

	var tlsConfig *tls.Config
	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("加载证书失败: %w", err)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	return &Server{
		config:    config,
		logger:    logger,
		handler:   handler,
		tlsConfig: tlsConfig,
	}, nil
}

// Start 启动服务器
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// 注册路由
	mux.HandleFunc("/", s.handler.HandleRoot)
	mux.HandleFunc("/v1", s.handler.HandleV1Root)
	mux.HandleFunc("/v1/models", s.handler.HandleModels)
	mux.HandleFunc("/v1/chat/completions", s.handler.HandleChatCompletions)

	server := &http.Server{
		Addr:      fmt.Sprintf(":%d", s.config.Server.Port),
		Handler:   mux,
		TLSConfig: s.tlsConfig,
	}

	if s.logger != nil {
		s.logger.Info("启动代理服务器，监听端口: %d", s.config.Server.Port)
		if len(s.config.APIs) > 0 {
			s.logger.Info("多后端配置已启用，共 %d 个API配置", len(s.config.APIs))
			for _, api := range s.config.APIs {
				status := "激活"
				if !api.Active {
					status = "未激活"
				}
				s.logger.Info("  - %s [%s]: %s -> %s", api.Name, status, api.Endpoint, api.CustomModelID)
			}
		}
	}

	if s.tlsConfig != nil {
		// 当使用TLSConfig时，certFile和keyFile可以为空，证书从TLSConfig中获取
		return server.ListenAndServeTLS("", "")
	}
	return server.ListenAndServe()
}

