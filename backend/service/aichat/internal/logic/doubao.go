package logic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ErizJ/JMall/backend/service/aichat/internal/config"
)

// doubaoMessage represents a chat message for the Doubao API.
type doubaoMessage struct {
	Role       string          `json:"role"`
	Content    string          `json:"content,omitempty"`
	ToolCalls  json.RawMessage `json:"tool_calls,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
}

// doubaoRequest is the request body for the Doubao chat completions API.
type doubaoRequest struct {
	Model    string            `json:"model"`
	Messages []doubaoMessage   `json:"messages"`
	Tools    []json.RawMessage `json:"tools,omitempty"`
	Stream   bool              `json:"stream"`
}

// doubaoChoice represents a single choice in the response.
type doubaoChoice struct {
	Index        int           `json:"index"`
	Message      doubaoMessage `json:"message"`
	Delta        doubaoMessage `json:"delta"`
	FinishReason string        `json:"finish_reason"`
}

// doubaoResponse is the response from the Doubao API.
type doubaoResponse struct {
	ID      string         `json:"id"`
	Choices []doubaoChoice `json:"choices"`
}

// doubaoToolCall parsed from the model response.
type doubaoToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

// callDoubao sends a non-streaming request to the Doubao API.
func callDoubao(cfg config.DoubaoConfig, req doubaoRequest) (*doubaoResponse, error) {
	req.Stream = false
	body, _ := json.Marshal(req)

	httpReq, err := http.NewRequest("POST", cfg.BaseUrl+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+cfg.ApiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("doubao API error %d: %s", resp.StatusCode, string(b))
	}

	var result doubaoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result, nil
}

// streamDoubao sends a streaming request and writes SSE events to the writer.
// It returns the full accumulated content and any tool_calls from the stream.
func streamDoubao(cfg config.DoubaoConfig, req doubaoRequest, w http.ResponseWriter, flusher http.Flusher) (string, []doubaoToolCall, error) {
	req.Stream = true
	body, _ := json.Marshal(req)

	httpReq, err := http.NewRequest("POST", cfg.BaseUrl+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+cfg.ApiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("doubao API error %d: %s", resp.StatusCode, string(b))
	}

	scanner := bufio.NewScanner(resp.Body)
	var fullContent strings.Builder
	var toolCalls []doubaoToolCall
	// Track partial tool call arguments across chunks
	toolCallArgBuilders := make(map[int]*strings.Builder)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk doubaoResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		for _, choice := range chunk.Choices {
			// Handle text content
			if choice.Delta.Content != "" {
				fullContent.WriteString(choice.Delta.Content)
				// Forward SSE to client
				sseData, _ := json.Marshal(map[string]string{"content": choice.Delta.Content})
				fmt.Fprintf(w, "data: %s\n\n", sseData)
				flusher.Flush()
			}

			// Handle tool calls in streaming mode
			if choice.Delta.ToolCalls != nil {
				var deltaToolCalls []struct {
					Index    int    `json:"index"`
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				}
				if err := json.Unmarshal(choice.Delta.ToolCalls, &deltaToolCalls); err == nil {
					for _, dtc := range deltaToolCalls {
						idx := dtc.Index
						if _, exists := toolCallArgBuilders[idx]; !exists {
							toolCallArgBuilders[idx] = &strings.Builder{}
							tc := doubaoToolCall{ID: dtc.ID, Type: dtc.Type}
							tc.Function.Name = dtc.Function.Name
							toolCalls = append(toolCalls, tc)
						}
						toolCallArgBuilders[idx].WriteString(dtc.Function.Arguments)
						if dtc.Function.Name != "" && toolCalls[idx].Function.Name == "" {
							toolCalls[idx].Function.Name = dtc.Function.Name
						}
						if dtc.ID != "" && toolCalls[idx].ID == "" {
							toolCalls[idx].ID = dtc.ID
						}
					}
				}
			}
		}
	}

	// Assemble final tool call arguments
	for idx, builder := range toolCallArgBuilders {
		if idx < len(toolCalls) {
			toolCalls[idx].Function.Arguments = builder.String()
		}
	}

	return fullContent.String(), toolCalls, nil
}
