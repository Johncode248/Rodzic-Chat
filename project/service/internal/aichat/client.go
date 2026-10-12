package aichat

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Johncode248/Rodzic-Chat/variables"
)

var ErrNotConfigured = errors.New("AI assistant is not configured yet -- missing ANTHROPIC_API_KEY")

const anthropicMessagesURL = "https://api.anthropic.com/v1/messages"
const anthropicVersion = "2023-06-01"

// contentBlock covers every block shape this app sends/receives on the Messages API: plain text
// (both turns), tool_use (assistant asking to call a tool), and tool_result (this app answering
// back with what the tool returned). Anthropic's API only requires the fields relevant to a
// block's own Type to be set -- the rest stay zero/omitted via `omitempty`.
type contentBlock struct {
	Type string `json:"type"`

	// text blocks
	Text string `json:"text,omitempty"`

	// tool_use blocks (assistant -> us)
	Id    string          `json:"id,omitempty"`
	Name  string          `json:"name,omitempty"`
	Input json.RawMessage `json:"input,omitempty"`

	// tool_result blocks (us -> assistant)
	ToolUseId string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`

	// image blocks (us -> assistant only, see project/service/internal/lessonplan) -- a photo of
	// a paper plan zajęć, base64-inlined per Anthropic's Messages API image content spec.
	Source *imageSource `json:"source,omitempty"`
}

type imageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type anthropicMessage struct {
	Role    string         `json:"role"`
	Content []contentBlock `json:"content"`
}

// toolDef is one entry in the Messages API's `tools` array -- InputSchema is a JSON Schema
// object (see tools.go for the ones this app declares).
type toolDef struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
	Tools     []toolDef          `json:"tools,omitempty"`
}

type anthropicUsage struct {
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
}

// TotalTokens sums input and output tokens for one round -- Anthropic meters them separately
// (and at different prices); this endpoint is currently hard-disabled server-side for MVP (see
// project/service/transport/aichat.Send), so nothing enforces a cap against this any more.
func (u anthropicUsage) TotalTokens() int64 {
	return u.InputTokens + u.OutputTokens
}

type anthropicResponse struct {
	Content    []contentBlock `json:"content"`
	StopReason string         `json:"stop_reason"`
	Usage      anthropicUsage `json:"usage"`
	Error      *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// anthropicHttpClient: a tool-use conversation can take several sequential round trips to
// Claude (one per tool-call round) on top of the NFZ/OSM calls each round makes -- 30s per
// individual Claude call still applies, but see converseWithTools for the round cap that bounds
// the total.
var anthropicHttpClient = &http.Client{Timeout: 30 * time.Second}

// callAnthropicOnce sends one Messages API request (the conversation so far, plus tool
// declarations if any) and returns the raw response for converseWithTools to interpret --
// unlike the old single-shot callAnthropic, this never decides on its own whether the
// conversation is "done"; a tool_use stop_reason is just as valid a return here as text.
func callAnthropicOnce(ctx context.Context, systemPrompt string, messages []anthropicMessage, tools []toolDef) (anthropicResponse, error) {
	if variables.AnthropicApiKey == "" {
		return anthropicResponse{}, ErrNotConfigured
	}

	body, err := json.Marshal(anthropicRequest{
		Model:     variables.AnthropicModel,
		MaxTokens: 1024,
		System:    systemPrompt,
		Messages:  messages,
		Tools:     tools,
	})
	if err != nil {
		return anthropicResponse{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, anthropicMessagesURL, bytes.NewReader(body))
	if err != nil {
		return anthropicResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", variables.AnthropicApiKey)
	req.Header.Set("anthropic-version", anthropicVersion)

	resp, err := anthropicHttpClient.Do(req)
	if err != nil {
		return anthropicResponse{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return anthropicResponse{}, err
	}

	var parsed anthropicResponse
	if err = json.Unmarshal(respBody, &parsed); err != nil {
		return anthropicResponse{}, fmt.Errorf("aichat: unexpected Anthropic response: %s", string(respBody))
	}

	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil {
			return anthropicResponse{}, fmt.Errorf("aichat: Anthropic error (%s): %s", parsed.Error.Type, parsed.Error.Message)
		}
		return anthropicResponse{}, fmt.Errorf("aichat: Anthropic request failed with status %d", resp.StatusCode)
	}

	return parsed, nil
}
