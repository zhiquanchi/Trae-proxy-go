package proxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// StreamResponse 处理流式响应转发
// 注意: customModelID 参数保留以保持接口一致性，但在流式响应中不进行模型ID替换
// 因为流式响应是逐块发送的，替换模型ID需要复杂的JSON解析，且Python版本也未实现
func StreamResponse(w http.ResponseWriter, r io.Reader, customModelID string) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("响应写入器不支持刷新")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// 直接转发流式数据
	buf := make([]byte, 4096)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				return fmt.Errorf("写入响应失败: %w", writeErr)
			}
			flusher.Flush()
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取流数据失败: %w", err)
		}
	}

	return nil
}

// SimulateStream 将非流式响应模拟为流式响应
func SimulateStream(w http.ResponseWriter, responseJSON map[string]interface{}, customModelID string) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("响应写入器不支持刷新")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// 提取内容
	choices, ok := responseJSON["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return fmt.Errorf("无效的响应格式")
	}

	choice, ok := choices[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("无效的选择格式")
	}

	message, ok := choice["message"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("无效的消息格式")
	}

	content, ok := message["content"].(string)
	if !ok {
		return fmt.Errorf("无效的内容格式")
	}

	// 发送初始块
	initialChunk := map[string]interface{}{
		"id":      "chatcmpl-simulated",
		"object":  "chat.completion.chunk",
		"created": 1,
		"model":   customModelID,
		"choices": []map[string]interface{}{
			{
				"index":         0,
				"delta":         map[string]interface{}{"role": "assistant"},
				"finish_reason": nil,
			},
		},
	}
	sendSSEChunk(w, initialChunk)
	flusher.Flush()

	// 将内容分成多个块发送
	chunkSize := 4
	for i := 0; i < len(content); i += chunkSize {
		end := i + chunkSize
		if end > len(content) {
			end = len(content)
		}
		chunk := content[i:end]

		chunkData := map[string]interface{}{
			"id":      "chatcmpl-simulated",
			"object":  "chat.completion.chunk",
			"created": 1,
			"model":   customModelID,
			"choices": []map[string]interface{}{
				{
					"index":         0,
					"delta":         map[string]interface{}{"content": chunk},
					"finish_reason": nil,
				},
			},
		}
		sendSSEChunk(w, chunkData)
		flusher.Flush()
	}

	// 发送完成标记
	finalChunk := map[string]interface{}{
		"id":      "chatcmpl-simulated",
		"object":  "chat.completion.chunk",
		"created": 1,
		"model":   customModelID,
		"choices": []map[string]interface{}{
			{
				"index":         0,
				"delta":         map[string]interface{}{},
				"finish_reason": "stop",
			},
		},
	}
	sendSSEChunk(w, finalChunk)
	flusher.Flush()

	// 发送结束标记
	if _, err := w.Write([]byte("data: [DONE]\n\n")); err != nil {
		return fmt.Errorf("写入结束标记失败: %w", err)
	}
	flusher.Flush()

	return nil
}

// sendSSEChunk 发送SSE格式的数据块
func sendSSEChunk(w http.ResponseWriter, data map[string]interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	w.Write([]byte("data: "))
	w.Write(jsonData)
	w.Write([]byte("\n\n"))
}

