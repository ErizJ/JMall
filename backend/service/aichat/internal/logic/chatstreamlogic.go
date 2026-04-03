package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ErizJ/JMall/backend/service/aichat/internal/mcp"
	"github.com/ErizJ/JMall/backend/service/aichat/internal/svc"
	"github.com/ErizJ/JMall/backend/service/aichat/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type ChatStreamLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatStreamLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatStreamLogic {
	return &ChatStreamLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChatStreamLogic) ChatStream(req *types.ChatReq, w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	toolDefs := mcp.GetToolDefinitions()
	tools := make([]json.RawMessage, 0, len(toolDefs))
	for _, td := range toolDefs {
		b, _ := json.Marshal(td)
		tools = append(tools, b)
	}

	messages := []doubaoMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: req.Message},
	}

	cfg := l.svcCtx.Config.Doubao

	// First call: non-streaming to check for tool calls
	resp, err := callDoubao(cfg, doubaoRequest{
		Model:    cfg.Model,
		Messages: messages,
		Tools:    tools,
	})
	if err != nil {
		sseError(w, flusher, "调用 AI 服务失败")
		return
	}
	if len(resp.Choices) == 0 {
		sseError(w, flusher, "AI 服务返回为空")
		return
	}

	choice := resp.Choices[0]

	// Handle tool call rounds (up to 3)
	for i := 0; i < 3 && choice.Message.ToolCalls != nil && choice.FinishReason != "stop"; i++ {
		var toolCalls []doubaoToolCall
		if err := json.Unmarshal(choice.Message.ToolCalls, &toolCalls); err != nil {
			break
		}

		// Send thinking indicator
		thinkData, _ := json.Marshal(map[string]string{"thinking": "正在查询商品信息..."})
		fmt.Fprintf(w, "data: %s\n\n", thinkData)
		flusher.Flush()

		messages = append(messages, choice.Message)

		for _, tc := range toolCalls {
			result, execErr := mcp.ExecuteTool(l.ctx, l.svcCtx, tc.Function.Name, tc.Function.Arguments)
			if execErr != nil {
				result = `{"error":"` + execErr.Error() + `"}`
			}
			messages = append(messages, doubaoMessage{
				Role:       "tool",
				Content:    result,
				ToolCallID: tc.ID,
			})
		}

		// Call again to see if more tools needed
		resp, err = callDoubao(cfg, doubaoRequest{
			Model:    cfg.Model,
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			sseError(w, flusher, "调用 AI 服务失败")
			return
		}
		if len(resp.Choices) == 0 {
			break
		}
		choice = resp.Choices[0]
	}

	// Final streaming response (no tools, just text)
	messages = append(messages, doubaoMessage{Role: "assistant", Content: choice.Message.Content})
	// Remove the last assistant message and re-stream
	messages = messages[:len(messages)-1]

	_, _, err = streamDoubao(cfg, doubaoRequest{
		Model:    cfg.Model,
		Messages: messages,
	}, w, flusher)
	if err != nil {
		sseError(w, flusher, "流式响应失败")
		return
	}

	// Send done event
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func sseError(w http.ResponseWriter, flusher http.Flusher, msg string) {
	data, _ := json.Marshal(map[string]string{"error": msg})
	fmt.Fprintf(w, "data: %s\n\n", data)
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}
