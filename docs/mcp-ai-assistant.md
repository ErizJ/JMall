# JMall MCP 智能助手：原理与实现

> 本文档基于 JMall 项目真实代码，从 MCP 协议原理到 Function Calling 工程实现完整讲解。
> 技术栈：Go (go-zero) + 豆包大模型 (Doubao) + SSE 流式传输 + Vue.js
> 目标：读完这篇文档，面试中能把 AI 智能助手从协议到落地讲清楚。

---

## 一、这个系统解决什么问题？

传统电商搜索：用户输入"手机" → 返回一堆商品列表 → 用户自己筛选。

AI 智能助手：用户说"帮我找一款 1500 左右的手机，要拍照好的" → AI 理解意图 → 自动调用搜索工具 → 筛选匹配商品 → 用自然语言回答。

核心区别：**AI 能理解自然语言意图，并自主决定调用哪些工具来获取信息**。

这就是 MCP（Model Context Protocol）和 Function Calling 的价值。

---

## 二、核心概念：Function Calling 与 MCP

### 2.1 什么是 Function Calling？

大模型（LLM）本身只会"说话"，不会"做事"。它不能查数据库、不能调 API。

Function Calling 是让 LLM "长手" 的机制：
1. 你告诉 LLM："这里有几个工具可以用"（工具定义）
2. 用户提问后，LLM 分析问题，决定要不要调工具、调哪个
3. LLM 输出"我要调 search_products，参数是 keyword=手机"
4. 你的代码执行这个工具，把结果返回给 LLM
5. LLM 根据工具返回的数据，生成最终回答

```
用户: "有什么热门手机？"
         │
         ▼
LLM 思考: "用户想看热门手机，我应该先搜索手机，再看热门排行"
         │
         ▼
LLM 输出: tool_call: search_products(keyword="手机")
         │
         ▼
你的代码: 执行 SQL 查询，返回 [{id:1, name:"Redmi K30", price:1599}, ...]
         │
         ▼
LLM 收到工具结果，生成回答:
"为您推荐以下热门手机：
 1. Redmi K30 — ¥1599，120Hz流速屏
 2. 小米CC9 Pro — ¥2599，1亿像素
 ..."
```

### 2.2 什么是 MCP？

MCP（Model Context Protocol）是 Anthropic 提出的标准化协议，定义了 LLM 与外部工具交互的规范。

我们的实现借鉴了 MCP 的思想：
- **工具定义**：用 JSON Schema 描述每个工具的名称、功能、参数
- **工具调用**：LLM 输出结构化的调用请求（函数名 + 参数 JSON）
- **结果返回**：工具执行结果以 JSON 格式返回给 LLM

本质上，MCP 就是 Function Calling 的标准化。不同 LLM 厂商（OpenAI、Anthropic、豆包）的 Function Calling 接口略有不同，但核心思想一致。

### 2.3 面试话术

> "我们的 AI 助手基于 Function Calling 机制实现。核心思路是：定义一组工具（搜索商品、
> 查分类、查促销等），把工具的 JSON Schema 传给大模型。用户提问后，大模型自主决定
> 调用哪些工具，我们执行工具并把结果返回给大模型，大模型再根据数据生成自然语言回答。
> 整个过程支持多轮工具调用（最多 3 轮），并且用 SSE 实现流式输出。"

---

## 三、整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                     前端 (AiChat.vue)                        │
│  悬浮按钮 → 聊天窗口 → SSE 流式接收 → 逐字渲染              │
└──────────────────────┬──────────────────────────────────────┘
                       │ POST /aichat/stream (SSE)
                       ▼
┌──────────────────────────────────────────────────────────────┐
│                  aichat-service (Go)                          │
│                                                               │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              ChatStreamLogic                             │ │
│  │                                                          │ │
│  │  1. 构建 messages（system prompt + user message）        │ │
│  │  2. 附加 tool definitions（7 个工具的 JSON Schema）      │ │
│  │  3. 调用豆包 API（非流式，检查是否需要调工具）           │ │
│  │  4. 如果 LLM 返回 tool_calls：                          │ │
│  │     ├── 发送 SSE: {"thinking": "正在查询商品信息..."}    │ │
│  │     ├── 执行工具（查 DB）                                │ │
│  │     ├── 把工具结果追加到 messages                        │ │
│  │     └── 再次调用豆包（最多循环 3 轮）                    │ │
│  │  5. 最终回答：流式调用豆包，SSE 逐字推送给前端           │ │
│  └─────────────────────────────────────────────────────────┘ │
│                                                               │
│  ┌──────────────┐  ┌──────────────────────────────────────┐  │
│  │  MCP Tools   │  │  Doubao Client                       │  │
│  │  · 搜索商品  │  │  · callDoubao()   非流式调用         │  │
│  │  · 查分类    │  │  · streamDoubao() 流式调用           │  │
│  │  · 查详情    │  │  · SSE 解析 + tool_call 组装         │  │
│  │  · 查热门    │  │                                      │  │
│  │  · 查促销    │  │                                      │  │
│  │  · 查满减    │  │                                      │  │
│  └──────┬───────┘  └──────────────────────────────────────┘  │
│         │                                                     │
│  ┌──────▼───────┐                                            │
│  │   MySQL      │                                            │
│  │  product     │                                            │
│  │  category    │                                            │
│  │  combination │                                            │
│  └──────────────┘                                            │
└──────────────────────────────────────────────────────────────┘
                       │
                       ▼
┌──────────────────────────────────────────────────────────────┐
│                  豆包大模型 API                                │
│  https://ark.cn-beijing.volces.com/api/v3/chat/completions   │
│  模型: doubao-1-5-pro-256k                                   │
└──────────────────────────────────────────────────────────────┘
```

---

## 四、完整交互流程

### 4.1 一次完整的对话（带工具调用）

以用户问"有什么热门手机推荐？"为例：

```
时间线：

T0  用户输入 "有什么热门手机推荐？"
    │
T1  前端 fetch('/api/aichat/stream', {message: "有什么热门手机推荐？"})
    │
T2  后端构建 messages:
    │  [
    │    {role: "system", content: "你是 JMall 商城的 AI 智能购物助手..."},
    │    {role: "user",   content: "有什么热门手机推荐？"}
    │  ]
    │  附加 7 个工具定义
    │
T3  调用豆包 API（非流式）→ 豆包返回:
    │  {
    │    finish_reason: "tool_calls",
    │    message: {
    │      tool_calls: [
    │        {id: "call_1", function: {name: "search_products", arguments: '{"keyword":"手机"}'}},
    │        {id: "call_2", function: {name: "get_hot_products", arguments: '{"limit":5}'}}
    │      ]
    │    }
    │  }
    │
T4  后端发送 SSE → 前端显示 "正在查询商品信息..."
    │
T5  后端执行工具:
    │  search_products("手机") → SELECT * FROM product WHERE name LIKE '%手机%'
    │  get_hot_products(5)     → SELECT * FROM product ORDER BY product_hot DESC LIMIT 5
    │
T6  把工具结果追加到 messages:
    │  [
    │    {role: "system", ...},
    │    {role: "user", "有什么热门手机推荐？"},
    │    {role: "assistant", tool_calls: [...]},          ← LLM 的工具调用请求
    │    {role: "tool", tool_call_id: "call_1", content: '[{id:1,name:"Redmi K30",...}]'},
    │    {role: "tool", tool_call_id: "call_2", content: '[{id:9,name:"小米电视",...}]'}
    │  ]
    │
T7  再次调用豆包（非流式）→ 检查是否还需要调工具
    │  → finish_reason: "stop"（不需要了）
    │
T8  最终回答：调用豆包（流式）→ SSE 逐字推送:
    │  data: {"content": "为"}
    │  data: {"content": "您"}
    │  data: {"content": "推荐"}
    │  data: {"content": "以下"}
    │  data: {"content": "热门手机"}
    │  data: {"content": "：\n\n"}
    │  data: {"content": "1. **Redmi K30** — ¥1599"}
    │  ...
    │  data: [DONE]
    │
T9  前端逐字渲染完成
```

### 4.2 为什么工具调用阶段用非流式，最终回答用流式？

- 工具调用阶段：LLM 需要输出完整的 `tool_calls` JSON，流式传输会把 JSON 拆成碎片，解析复杂且容易出错。用非流式一次拿到完整结果更可靠。
- 最终回答阶段：纯文本输出，流式传输让用户看到"逐字打出"的效果，体验好。

这是一个工程上的务实选择。

---

## 五、MCP 工具定义

### 5.1 我们定义了 7 个工具

| 工具名 | 功能 | 参数 | 对应 SQL |
|--------|------|------|---------|
| `search_products` | 关键词搜索商品 | keyword | `WHERE name LIKE '%keyword%'` |
| `get_categories` | 获取所有分类 | 无 | `SELECT * FROM category` |
| `get_product_detail` | 商品详情 | product_id | `WHERE product_id = ?` |
| `get_products_by_category` | 按分类查商品 | category_id | `WHERE category_id = ?` |
| `get_hot_products` | 热门排行 | limit | `ORDER BY product_hot DESC` |
| `get_promotion_products` | 促销商品 | limit | `WHERE product_isPromotion > 0` |
| `get_combination_discounts` | 满减活动 | 无 | `SELECT * FROM combination_product` |

### 5.2 工具定义格式（JSON Schema）

每个工具用 JSON Schema 描述，LLM 根据这个 Schema 知道工具能做什么、需要什么参数：

```go
// tools.go
func GetToolDefinitions() []ToolDef {
    return []ToolDef{
        {
            Type: "function",
            Function: FunctionDef{
                Name:        "search_products",
                Description: "根据关键词搜索商品，返回商品名称、价格、库存等信息",
                Parameters: json.RawMessage(`{
                    "type": "object",
                    "properties": {
                        "keyword": {
                            "type": "string",
                            "description": "搜索关键词，如手机、电视等"
                        }
                    },
                    "required": ["keyword"]
                }`),
            },
        },
        // ... 其他 6 个工具
    }
}
```

**关键点：`description` 字段非常重要。** LLM 根据 description 判断什么时候该调这个工具。如果 description 写得不好，LLM 可能选错工具或者不调工具。

### 5.3 工具执行（路由分发）

```go
// tools.go
func ExecuteTool(ctx context.Context, svcCtx *svc.ServiceContext, name string, argsJSON string) (string, error) {
    switch name {
    case "search_products":
        return execSearchProducts(ctx, svcCtx, argsJSON)
    case "get_categories":
        return execGetCategories(ctx, svcCtx)
    case "get_product_detail":
        return execGetProductDetail(ctx, svcCtx, argsJSON)
    // ... 其他工具
    default:
        return "", fmt.Errorf("unknown tool: %s", name)
    }
}
```

每个工具的执行逻辑就是：解析参数 JSON → 查数据库 → 把结果序列化为 JSON 返回。

---

## 六、System Prompt 设计

System Prompt 定义了 AI 助手的"人设"和行为边界：

```go
const systemPrompt = `你是 JMall 商城的 AI 智能购物助手。你可以帮助用户：
1. 搜索和查询商品信息（名称、价格、库存等）
2. 查看商品分类
3. 了解热门商品和促销活动
4. 查询组合优惠和满减信息
5. 提供购物建议和商品推荐

请用友好、专业的语气回答用户问题。当需要查询商品信息时，请使用提供的工具函数。
回答时请使用中文，并尽量提供具体的商品信息（如价格、库存等）。
如果用户问的问题与购物无关，请礼貌地引导用户回到购物相关话题。`
```

设计要点：
- 明确角色定位（购物助手，不是通用 AI）
- 列出能力范围（5 项具体能力）
- 指定行为规范（友好、专业、中文、具体数据）
- 设定边界（非购物话题引导回来）

---

## 七、多轮工具调用机制

LLM 可能需要多次调用工具才能回答一个问题。比如：

用户："Redmi K30 和小米CC9 Pro 哪个更值得买？"

```
第 1 轮: LLM 调用 get_product_detail(product_id=1)  → 获取 Redmi K30 详情
第 2 轮: LLM 调用 get_product_detail(product_id=3)  → 获取 CC9 Pro 详情
第 3 轮: LLM 根据两个商品的数据，生成对比分析回答
```

代码中用循环实现，最多 3 轮：

```go
// chatlogic.go
for i := 0; i < 3; i++ {
    resp, _ := callDoubao(cfg, doubaoRequest{
        Model:    cfg.Model,
        Messages: messages,
        Tools:    tools,
    })

    choice := resp.Choices[0]

    // 如果不需要调工具了，直接返回
    if choice.Message.ToolCalls == nil || choice.FinishReason == "stop" {
        return &types.ChatResp{Code: "200", Reply: choice.Message.Content}, nil
    }

    // 解析并执行工具调用
    var toolCalls []doubaoToolCall
    json.Unmarshal(choice.Message.ToolCalls, &toolCalls)

    // 把 LLM 的工具调用请求加入 messages
    messages = append(messages, choice.Message)

    // 执行每个工具，把结果加入 messages
    for _, tc := range toolCalls {
        result, _ := mcp.ExecuteTool(l.ctx, l.svcCtx, tc.Function.Name, tc.Function.Arguments)
        messages = append(messages, doubaoMessage{
            Role:       "tool",
            Content:    result,
            ToolCallID: tc.ID,
        })
    }
    // 下一轮循环会带着工具结果再次调用 LLM
}
```

**messages 数组的演变过程：**

```
初始:
  [system, user]

第 1 轮后:
  [system, user, assistant(tool_calls), tool(result_1)]

第 2 轮后:
  [system, user, assistant(tool_calls), tool(result_1), assistant(tool_calls_2), tool(result_2)]

最终:
  LLM 看到所有工具结果，生成最终文本回答
```

这就是 Function Calling 的"对话上下文"机制——每一轮的工具调用和结果都作为上下文传给下一轮。

---

## 八、SSE 流式传输

### 8.1 什么是 SSE？

SSE（Server-Sent Events）是 HTTP 协议上的单向流式传输。服务端可以持续向客户端推送数据，不需要 WebSocket。

格式：
```
data: {"content": "你"}

data: {"content": "好"}

data: [DONE]
```

每条消息以 `data: ` 开头，以两个换行 `\n\n` 结尾。

### 8.2 后端实现

```go
// chatstreamlogic.go
func (l *ChatStreamLogic) ChatStream(req *types.ChatReq, w http.ResponseWriter, r *http.Request) {
    // 设置 SSE 响应头
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    flusher, _ := w.(http.Flusher)

    // ... 工具调用阶段（非流式）...

    // 工具调用时发送"思考中"提示
    thinkData, _ := json.Marshal(map[string]string{"thinking": "正在查询商品信息..."})
    fmt.Fprintf(w, "data: %s\n\n", thinkData)
    flusher.Flush()

    // ... 执行工具 ...

    // 最终回答：流式输出
    streamDoubao(cfg, doubaoRequest{
        Model:    cfg.Model,
        Messages: messages,
    }, w, flusher)

    fmt.Fprintf(w, "data: [DONE]\n\n")
    flusher.Flush()
}
```

### 8.3 豆包流式响应解析

豆包 API 的流式响应也是 SSE 格式，每个 chunk 包含一小段文本：

```go
// doubao.go — streamDoubao
func streamDoubao(cfg, req, w, flusher) (string, []doubaoToolCall, error) {
    scanner := bufio.NewScanner(resp.Body)
    var fullContent strings.Builder

    for scanner.Scan() {
        line := scanner.Text()
        if !strings.HasPrefix(line, "data: ") { continue }
        data := strings.TrimPrefix(line, "data: ")
        if data == "[DONE]" { break }

        var chunk doubaoResponse
        json.Unmarshal([]byte(data), &chunk)

        for _, choice := range chunk.Choices {
            if choice.Delta.Content != "" {
                fullContent.WriteString(choice.Delta.Content)
                // 转发给前端
                sseData, _ := json.Marshal(map[string]string{"content": choice.Delta.Content})
                fmt.Fprintf(w, "data: %s\n\n", sseData)
                flusher.Flush()
            }
        }
    }
}
```

数据流：`豆包 API → SSE → 后端解析 → SSE → 前端渲染`

后端是一个"SSE 中继"——从豆包接收 SSE，解析后再以 SSE 推送给前端。

### 8.4 前端 SSE 接收

```javascript
// AiChat.vue
async sendStreamMessage(msg) {
    const response = await fetch('/api/aichat/stream', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', ...authHeader },
        body: JSON.stringify({ message: msg }),
    })

    const reader = response.body.getReader()
    const decoder = new TextDecoder()
    let botMsg = { role: 'bot', content: '' }
    this.messages.push(botMsg)

    while (true) {
        const { done, value } = await reader.read()
        if (done) break

        // 解析 SSE 数据
        const lines = decoder.decode(value).split('\n')
        for (const line of lines) {
            if (!line.startsWith('data: ')) continue
            const parsed = JSON.parse(line.slice(6))

            if (parsed.thinking) {
                this.thinkingText = parsed.thinking  // 显示"正在查询..."
            } else if (parsed.content) {
                botMsg.content += parsed.content      // 逐字追加
                this.$set(this.messages, this.messages.length - 1, { ...botMsg })
            }
        }
    }
}
```

用户看到的效果：先显示"正在查询商品信息..."，然后文字逐字打出来。

---

## 九、流式中的 Tool Call 组装

这是一个工程难点。流式传输时，tool_calls 的参数 JSON 会被拆成多个 chunk：

```
chunk 1: tool_calls: [{index:0, id:"call_1", function:{name:"search_products", arguments:'{"key'}}]
chunk 2: tool_calls: [{index:0, function:{arguments:'word":"手'}}]
chunk 3: tool_calls: [{index:0, function:{arguments:'机"}'}}]
```

需要按 index 把碎片拼接起来：

```go
// doubao.go
toolCallArgBuilders := make(map[int]*strings.Builder)

for _, dtc := range deltaToolCalls {
    idx := dtc.Index
    if _, exists := toolCallArgBuilders[idx]; !exists {
        toolCallArgBuilders[idx] = &strings.Builder{}
        // 第一个 chunk 包含 id 和 name
        toolCalls = append(toolCalls, doubaoToolCall{ID: dtc.ID, Type: dtc.Type})
        toolCalls[idx].Function.Name = dtc.Function.Name
    }
    // 每个 chunk 的 arguments 追加到 builder
    toolCallArgBuilders[idx].WriteString(dtc.Function.Arguments)
}

// 最后组装完整的 arguments
for idx, builder := range toolCallArgBuilders {
    toolCalls[idx].Function.Arguments = builder.String()
}
```

---

## 十、前端交互设计

### 10.1 悬浮按钮 + 聊天窗口

```
┌──────────────────────────────────┐
│  JMall 智能助手              ─   │  ← 标题栏
├──────────────────────────────────┤
│                                  │
│  🤖 你好！我是 JMall 智能购物助手│  ← 欢迎消息
│     🔥 热门商品推荐              │  ← 快捷建议（可点击）
│     🏷️ 促销活动                  │
│     📱 搜索商品                  │
│                                  │
│  👤 有什么热门手机推荐？         │  ← 用户消息
│                                  │
│  🤖 正在查询商品信息...          │  ← 思考中提示
│                                  │
│  🤖 为您推荐以下热门手机：       │  ← AI 回答（逐字打出）
│     1. Redmi K30 — ¥1599        │
│     2. 小米CC9 Pro — ¥2599      │
│                                  │
├──────────────────────────────────┤
│  [输入你的问题...        ] [发送]│  ← 输入框
└──────────────────────────────────┘
                                    [💬]  ← 悬浮按钮
```

### 10.2 Mock 模式

开发时可能没有豆包 API Key，前端支持 Mock 模式：

```javascript
isMock: process.env.VUE_APP_USE_MOCK === 'true'

if (this.isMock) {
    await this.sendMockMessage(msg)   // 走 Axios，被 mock 拦截器捕获
} else {
    await this.sendStreamMessage(msg) // 走 fetch SSE，真实调用
}
```

Mock 模式用 Axios 调非流式接口 `/aichat/chat`，正常模式用 fetch 调流式接口 `/aichat/stream`。

### 10.3 为什么流式用 fetch 而不是 Axios？

Axios 不支持流式读取响应体。`fetch` 的 `response.body.getReader()` 可以逐块读取 SSE 数据。这是浏览器原生 API 的能力。

---

## 十一、消息协议（messages 数组）

整个对话的核心数据结构是 `messages` 数组，它记录了完整的对话上下文：

```json
[
  {"role": "system",    "content": "你是 JMall 商城的 AI 智能购物助手..."},
  {"role": "user",      "content": "有什么热门手机推荐？"},
  {"role": "assistant", "tool_calls": [{"id":"call_1", "function":{"name":"search_products","arguments":"{\"keyword\":\"手机\"}"}}]},
  {"role": "tool",      "tool_call_id": "call_1", "content": "[{\"id\":1,\"name\":\"Redmi K30\"}]"},
  {"role": "assistant", "content": "为您推荐以下热门手机：..."}
]
```

四种角色：
| 角色 | 说明 |
|------|------|
| `system` | 系统提示词，定义 AI 的人设和行为规范 |
| `user` | 用户的输入 |
| `assistant` | AI 的回复（可能是文本，也可能是 tool_calls） |
| `tool` | 工具执行结果，必须带 `tool_call_id` 关联到对应的调用 |

`tool_call_id` 是关键——它把工具结果和对应的调用请求关联起来。如果一次调用了多个工具，每个结果都要有正确的 `tool_call_id`。

---

## 十二、错误处理

### 12.1 豆包 API 调用失败

```go
if resp.StatusCode != http.StatusOK {
    b, _ := io.ReadAll(resp.Body)
    return nil, fmt.Errorf("doubao API error %d: %s", resp.StatusCode, string(b))
}
```

失败时通过 SSE 发送错误消息给前端：

```go
func sseError(w http.ResponseWriter, flusher http.Flusher, msg string) {
    data, _ := json.Marshal(map[string]string{"error": msg})
    fmt.Fprintf(w, "data: %s\n\n", data)
    fmt.Fprintf(w, "data: [DONE]\n\n")
    flusher.Flush()
}
```

### 12.2 工具执行失败

```go
result, execErr := mcp.ExecuteTool(l.ctx, l.svcCtx, tc.Function.Name, tc.Function.Arguments)
if execErr != nil {
    result = `{"error":"` + execErr.Error() + `"}`
}
// 即使工具失败，也把错误信息返回给 LLM，让它自己处理
messages = append(messages, doubaoMessage{Role: "tool", Content: result, ToolCallID: tc.ID})
```

工具失败不会中断对话——错误信息作为工具结果返回给 LLM，LLM 会说"抱歉，查询失败了"之类的话。

### 12.3 前端错误处理

```javascript
try {
    // ... SSE 读取 ...
} catch (err) {
    this.messages.push({ role: 'bot', content: '网络异常，请稍后再试。' })
} finally {
    this.loading = false
    this.thinkingText = ''
}
```

---

## 十三、面试高频问题

### Q1：什么是 Function Calling？和普通 API 调用有什么区别？

Function Calling 是让 LLM 自主决定调用哪些外部工具的机制。普通 API 调用是程序员写死的"if-else"逻辑，Function Calling 是 LLM 根据用户意图动态决定的。比如用户问"手机多少钱"，LLM 自己判断需要调 search_products 工具，而不是程序员写正则匹配"手机"关键词。

### Q2：MCP 和 Function Calling 是什么关系？

MCP 是 Anthropic 提出的标准化协议，定义了 LLM 与外部工具交互的规范。Function Calling 是各 LLM 厂商（OpenAI、豆包等）的具体实现。MCP 是标准，Function Calling 是实现。我们的代码借鉴了 MCP 的工具定义格式（JSON Schema），但调用的是豆包的 Function Calling API。

### Q3：为什么工具调用阶段用非流式，最终回答用流式？

工具调用阶段 LLM 输出的是结构化的 JSON（tool_calls），流式传输会把 JSON 拆成碎片，解析复杂且容易出错。最终回答是纯文本，流式传输让用户看到逐字打出的效果，体验好。这是工程上的务实选择。

### Q4：如果 LLM 调了错误的工具怎么办？

两层保障：一是工具的 description 写得足够清晰，减少误调；二是工具执行失败会返回错误 JSON 给 LLM，LLM 会自行修正或告知用户。最多 3 轮循环也是一个安全阀——防止 LLM 无限循环调工具。

### Q5：为什么最多只允许 3 轮工具调用？

防止无限循环。如果 LLM 出现幻觉，可能会不断调用工具。3 轮足够覆盖绝大多数场景（搜索→详情→对比），超过 3 轮说明问题太复杂或 LLM 出了问题，直接返回兜底回答。

### Q6：SSE 和 WebSocket 有什么区别？为什么选 SSE？

SSE 是单向的（服务端→客户端），基于 HTTP，实现简单。WebSocket 是双向的，需要额外的握手和连接管理。AI 对话场景是"请求-响应"模式，不需要双向通信，SSE 完全够用且更轻量。

### Q7：流式传输中 tool_calls 的参数被拆成碎片怎么处理？

用 `map[int]*strings.Builder` 按 index 追踪每个 tool_call 的参数碎片。每个 chunk 到达时，把 arguments 追加到对应 index 的 builder。所有 chunk 接收完后，从 builder 中取出完整的 arguments JSON。

### Q8：System Prompt 怎么设计的？有什么讲究？

四个要素：角色定位（购物助手）、能力范围（5 项具体能力）、行为规范（友好、中文、具体数据）、边界约束（非购物话题引导回来）。description 越具体，LLM 的行为越可控。

### Q9：如果要支持多轮对话（记住上下文），怎么做？

当前实现是单轮对话（每次请求只有 system + user）。要支持多轮，需要在前端维护 messages 历史，每次请求把完整的对话历史发给后端。后端直接把历史 messages 传给 LLM。注意要控制 messages 长度，超过 token 限制时截断早期消息。

### Q10：这个架构能换成其他大模型吗？

可以。`doubao.go` 封装了所有 LLM 调用逻辑，只需要改 `callDoubao` 和 `streamDoubao` 两个函数的 HTTP 请求格式。OpenAI、Claude、通义千问的 Function Calling 接口格式大同小异，核心的 messages + tools 结构是通用的。

---

## 十四、项目结构

```
backend/service/aichat/
├── aichat.go                          # 入口（启用 CORS）
├── etc/aichat-api.yaml                # 配置（豆包 API Key、模型名、Base URL）
└── internal/
    ├── config/config.go               # 配置结构体（含 DoubaoConfig）
    ├── handler/
    │   ├── routes.go                  # 2 个接口：/aichat/chat + /aichat/stream
    │   ├── chathandler.go             # 非流式接口
    │   └── chatstreamhandler.go       # 流式接口（SSE）
    ├── logic/
    │   ├── chatlogic.go               # 非流式对话（工具调用循环 + 最终回答）
    │   ├── chatstreamlogic.go         # 流式对话（工具调用 + SSE 推送）
    │   └── doubao.go                  # 豆包 API 封装（callDoubao + streamDoubao）
    ├── mcp/
    │   └── tools.go                   # 7 个工具定义 + 执行路由 + DB 查询
    ├── svc/servicecontext.go          # 依赖注入（ProductModel、CategoryModel 等）
    └── types/types.go                 # ChatReq / ChatResp

frontend/src/components/
└── AiChat.vue                         # 悬浮聊天窗口（SSE 接收 + 逐字渲染 + Mock 模式）
```

---

## 十五、一句话总结

MCP 智能助手的核心是 **Function Calling 循环**：把工具定义（JSON Schema）传给大模型，大模型根据用户意图自主决定调用哪些工具，我们执行工具并把结果返回，大模型再根据数据生成自然语言回答。工具调用阶段用非流式保证 JSON 完整性，最终回答用 SSE 流式传输实现逐字打出效果。整个架构是 LLM 无关的——换模型只需改 HTTP 调用层，工具定义和执行逻辑完全复用。
