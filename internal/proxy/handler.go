package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"trae-proxy-go/internal/logger"
	"trae-proxy-go/pkg/models"
)

// Handler 处理器结构
type Handler struct {
	config *models.Config
	logger *logger.Logger
}

// NewHandler 创建新的处理器
func NewHandler(config *models.Config, logger *logger.Logger) *Handler {
	return &Handler{
		config: config,
		logger: logger,
	}
}

// HandleRoot 处理根路径
func (h *Handler) HandleRoot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]string{
		"message": "Welcome to the OpenAI API! Documentation is available at https://platform.openai.com/docs/api-reference",
	}
	h.writeJSON(w, response)
}

// HandleV1Root 处理/v1路径
func (h *Handler) HandleV1Root(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"message": "OpenAI API v1 endpoint",
		"endpoints": map[string]string{
			"chat/completions": "/v1/chat/completions",
		},
	}
	h.writeJSON(w, response)
}

// HandleModels 处理模型列表请求
func (h *Handler) HandleModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	models := []map[string]interface{}{}
	for _, api := range h.config.APIs {
		if api.Active {
			models = append(models, map[string]interface{}{
				"id":        api.CustomModelID,
				"object":    "model",
				"created":   1,
				"owned_by":  "trae-proxy",
			})
		}
	}

	response := map[string]interface{}{
		"object": "list",
		"data":   models,
	}
	h.writeJSON(w, response)
}

// HandleChatCompletions 处理聊天完成请求
func (h *Handler) HandleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	// 检查Content-Type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		h.writeError(w, "Content-Type必须为application/json", http.StatusBadRequest)
		return
	}

	// 解析请求体
	var reqJSON map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&reqJSON); err != nil {
		h.writeError(w, fmt.Sprintf("无效的JSON请求体: %v", err), http.StatusBadRequest)
		return
	}

	// 调试日志
	if h.logger != nil {
		h.logger.Debug("请求头: %v", r.Header)
		reqJSONBytes, _ := json.Marshal(reqJSON)
		h.logger.Debug("请求体: %s", string(reqJSONBytes))
	}

	// 获取请求的模型ID
	requestedModel, _ := reqJSON["model"].(string)

	// 选择后端API
	selectedBackend := selectBackendByModel(h.config, requestedModel)
	if selectedBackend == nil {
		h.writeError(w, "未找到可用的后端API配置", http.StatusInternalServerError)
		return
	}

	targetAPIURL := selectedBackend.Endpoint
	targetModelID := selectedBackend.TargetModelID
	customModelID := selectedBackend.CustomModelID
	streamMode := selectedBackend.StreamMode

	if h.logger != nil {
		h.logger.Info("选择后端: %s -> %s", selectedBackend.Name, targetAPIURL)
	}

	// 修改模型ID
	reqJSON["model"] = targetModelID

	// 处理流模式
	// streamMode: "true" 强制开启, "false" 强制关闭, "" 或不设置则保持原请求设置
	if streamMode == "true" {
		reqJSON["stream"] = true
	} else if streamMode == "false" {
		reqJSON["stream"] = false
	}
	// 如果streamMode为空，保持原请求的stream设置（不修改）

	// 准备转发请求
	reqBody, err := json.Marshal(reqJSON)
	if err != nil {
		h.writeError(w, fmt.Sprintf("序列化请求失败: %v", err), http.StatusInternalServerError)
		return
	}

	targetURL := fmt.Sprintf("%s/v1/chat/completions", targetAPIURL)
	if h.logger != nil {
		h.logger.Debug("转发请求到: %s", targetURL)
	}

	// 创建转发请求
	req, err := http.NewRequest(http.MethodPost, targetURL, bytes.NewBuffer(reqBody))
	if err != nil {
		h.writeError(w, fmt.Sprintf("创建请求失败: %v", err), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	if authHeader := r.Header.Get("Authorization"); authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		if h.logger != nil {
			h.logger.Error("请求失败: %v", err)
		}
		h.writeError(w, fmt.Sprintf("请求异常: %v", err), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// 处理错误响应
	if resp.StatusCode >= 400 {
		errorBody, _ := io.ReadAll(resp.Body)
		var errorJSON map[string]interface{}
		if err := json.Unmarshal(errorBody, &errorJSON); err == nil {
			h.writeJSON(w, errorJSON, resp.StatusCode)
		} else {
			h.writeError(w, fmt.Sprintf("HTTP错误: %s", resp.Status), resp.StatusCode)
		}
		return
	}

	// 检查是否为流式响应
	isStream, _ := reqJSON["stream"].(bool)
	if isStream {
		// 流式响应
		if h.logger != nil {
			h.logger.Debug("返回流式响应")
		}
		if err := StreamResponse(w, resp.Body, customModelID); err != nil {
			if h.logger != nil {
				h.logger.Error("流式响应处理失败: %v", err)
			}
		}
		return
	}

	// 非流式响应
	var responseJSON map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseJSON); err != nil {
		h.writeError(w, fmt.Sprintf("解析响应失败: %v", err), http.StatusInternalServerError)
		return
	}

	if h.logger != nil {
		responseJSONBytes, _ := json.Marshal(responseJSON)
		h.logger.Debug("响应体: %s", string(responseJSONBytes))
	}

	// 修改响应中的模型ID
	if responseJSON["model"] != nil {
		responseJSON["model"] = customModelID
	}

	h.writeJSON(w, responseJSON)
}

// writeJSON 写入JSON响应
func (h *Handler) writeJSON(w http.ResponseWriter, data interface{}, statusCode ...int) {
	w.Header().Set("Content-Type", "application/json")
	code := http.StatusOK
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

// writeError 写入错误响应
func (h *Handler) writeError(w http.ResponseWriter, message string, statusCode int) {
	h.writeJSON(w, map[string]string{"error": message}, statusCode)
}

