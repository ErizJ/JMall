package logic

import (
	"context"
	"encoding/json"

	"github.com/ErizJ/JMall/backend/service/aichat/internal/mcp"
	"github.com/ErizJ/JMall/backend/service/aichat/internal/svc"
	"github.com/ErizJ/JMall/backend/service/aichat/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

const systemPrompt = `你是 JMall 商城的 AI 智能购物助手。你可以帮助用户：
1. 搜索和查询商品信息（名称、价格、库存等）
2. 查看商品分类
3. 了解热门商品和促销活动
4. 查询组合优惠和满减信息
5. 提供购物建议和商品推荐

请用友好、专业的语气回答用户问题。当需要查询商品信息时，请使用提供的工具函数。
回答时请使用中文，并尽量提供具体的商品信息（如价格、库存等）。
如果用户问的问题与购物无关，请礼貌地引导用户回到购物相关话题。`

type ChatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewChatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ChatLogic {
	return &ChatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ChatLogic) Chat(req *types.ChatReq) (*types.ChatResp, error) {
	// Build tool definitions as raw JSON
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

	// Call Doubao with tools, handle up to 3 rounds of tool calls
	cfg := l.svcCtx.Config.Doubao
	for i := 0; i < 3; i++ {
		resp, err := callDoubao(cfg, doubaoRequest{
			Model:    cfg.Model,
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			return nil, err
		}
		if len(resp.Choices) == 0 {
			return &types.ChatResp{Code: "200", Reply: "抱歉，我暂时无法回答您的问题。"}, nil
		}

		choice := resp.Choices[0]

		// If no tool calls, return the content directly
		if choice.Message.ToolCalls == nil || choice.FinishReason == "stop" {
			return &types.ChatResp{Code: "200", Reply: choice.Message.Content}, nil
		}

		// Parse and execute tool calls
		var toolCalls []doubaoToolCall
		if err := json.Unmarshal(choice.Message.ToolCalls, &toolCalls); err != nil {
			return &types.ChatResp{Code: "200", Reply: choice.Message.Content}, nil
		}

		// Add assistant message with tool calls
		messages = append(messages, choice.Message)

		// Execute each tool call and add results
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
	}

	return &types.ChatResp{Code: "200", Reply: "抱歉，处理您的请求时遇到了问题，请稍后再试。"}, nil
}
