package ernieclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/sayerxofficial/langchaingo/llms"
	"github.com/sayerxofficial/langchaingo/llms/streaming"
)

const (
	defaultBaseURL = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1"
	streamStopFlag = "\"is_end\": true"
)

// ChatRequest is a request to complete a chat completion..
type ChatRequest struct {
	Model            string         `json:"model,omitempty"`
	Messages         []*ChatMessage `json:"messages"`
	Temperature      float64        `json:"temperature"`
	TopP             float64        `json:"top_p,omitempty"`
	MaxTokens        int            `json:"max_tokens,omitempty"`
	N                int            `json:"n,omitempty"`
	StopWords        []string       `json:"stop,omitempty"`
	Stream           bool           `json:"stream,omitempty"`
	FrequencyPenalty float64        `json:"frequency_penalty,omitempty"`
	PresencePenalty  float64        `json:"presence_penalty,omitempty"`

	// If the 'functions' parameter is set, setting the 'system' parameter is not supported.
	System string `json:"system,omitempty"`

	// Function definitions to include in the request.
	Functions []FunctionDefinition `json:"functions,omitempty"`
	// FunctionCallBehavior is the behavior to use when calling functions.
	//
	// If a specific function should be invoked, use the format:
	// `{"name": "my_function"}`
	FunctionCallBehavior FunctionCallBehavior `json:"function_call,omitempty"`

	// StreamingFunc is a function to be called for each chunk of a streaming response.
	// Return an error to stop streaming early.
	StreamingFunc streaming.Callback `json:"-"`
}

// ChatMessage is a message in a chat request.
type ChatMessage struct {
	// The role of the author of this message. One of system, user, or assistant.
	Role string `json:"role"`
	// The content of the message.
	Content string `json:"content"`
	// The name of the author of this message. May contain a-z, A-Z, 0-9, and underscores,
	// with a maximum length of 64 characters.
	Name string `json:"name,omitempty"`

	// FunctionCall represents a function call to be made in the message.
	FunctionCall *llms.FunctionCall `json:"function_call,omitempty"`
}

// ChatChoice is a choice in a chat response.
type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// ChatUsage is the usage of a chat completion request.
type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type ChatResponse struct {
	ID               string           `json:"id"`
	Object           string           `json:"object"`
	Created          int              `json:"created"`
	Result           string           `json:"result"`
	IsTruncated      bool             `json:"is_truncated"`
	NeedClearHistory bool             `json:"need_clear_history"`
	FunctionCall     *FunctionCallRes `json:"function_call,omitempty"`
	Usage            struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

type FunctionCallRes struct {
	Name      string `json:"name"`
	Thoughts  string `json:"thoughts"`
	Arguments string `json:"arguments"`
}

type StreamedChatResponsePayload struct {
	ID               string           `json:"id"`
	Object           string           `json:"object"`
	Created          int              `json:"created"`
	SentenceID       int              `json:"sentence_id"`
	IsEnd            bool             `json:"is_end"`
	IsTruncated      bool             `json:"is_truncated"`
	Result           string           `json:"result"`
	NeedClearHistory bool             `json:"need_clear_history"`
	FunctionCall     *FunctionCallRes `json:"function_call,omitempty"`
	Usage            struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// FunctionDefinition is a definition of a function that can be called by the model.
type FunctionDefinition struct {
	// Name is the name of the function.
	Name string `json:"name"`
	// Description is a description of the function.
	Description string `json:"description"`
	// Parameters is a list of parameters for the function.
	Parameters any `json:"parameters"`
}

// FunctionCallBehavior is the behavior to use when calling functions.
type FunctionCallBehavior string

const (
	// FunctionCallBehaviorUnspecified is the empty string.
	FunctionCallBehaviorUnspecified FunctionCallBehavior = ""
	// FunctionCallBehaviorNone will not call any functions.
	FunctionCallBehaviorNone FunctionCallBehavior = "none"
	// FunctionCallBehaviorAuto will call functions automatically.
	FunctionCallBehaviorAuto FunctionCallBehavior = "auto"
)

// FunctionCall is a call to a function.
type FunctionCall struct {
	// Name is the name of the function to call.
	Name string `json:"name"`
	// Arguments is the set of arguments to pass to the function.
	Arguments string `json:"arguments"`
}

func (c *Client) createChat(ctx context.Context, payload *ChatRequest) (*ChatResponse, error) {
	if payload.StreamingFunc != nil {
		payload.Stream = true
	}
	// Build request payload
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	// Build request
	body := bytes.NewReader(payloadBytes)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.buildURL(c.ModelPath), body)
	if err != nil {
		return nil, err
	}

	c.setHeaders(req)

	// Send request
	r, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("API returned unexpected status code: %d", r.StatusCode)

		// No need to check the error here: if it fails, we'll just return the
		// status code.
		var errResp errorMessage
		if err := json.NewDecoder(r.Body).Decode(&errResp); err != nil {
			return nil, errors.New(msg)
		}

		return nil, fmt.Errorf("%s: %s", msg, errResp.Error.Message)
	}
	if payload.StreamingFunc != nil {
		return parseStreamingChatResponse(ctx, r, payload)
	}
	// Parse response
	var response ChatResponse
	return &response, json.NewDecoder(r.Body).Decode(&response)
}

func parseStreamingChatResponse(ctx context.Context, r *http.Response, payload *ChatRequest) (*ChatResponse, error) { //nolint:cyclop,lll
	scanner := bufio.NewScanner(r.Body)
	responseChan := make(chan StreamedChatResponsePayload)
	go func() {
		defer close(responseChan)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "data:") {
				log.Fatalf("unexpected line: %v", line)
			}
			data := strings.TrimPrefix(line, "data: ")
			var streamPayload StreamedChatResponsePayload
			err := json.NewDecoder(bytes.NewReader([]byte(data))).Decode(&streamPayload)
			if err != nil {
				log.Fatalf("failed to decode stream payload: %v", err)
			}
			responseChan <- streamPayload
			if strings.Contains(data, streamStopFlag) {
				return
			}
		}
		if err := scanner.Err(); err != nil {
			log.Println("issue scanning response:", err)
		}
	}()
	// Parse response
	response := ChatResponse{}

	toolCallID := 0 // try to keep conversation of tool calls id existing
	incToolCallID := func(args string) {
		var data any
		if err := json.Unmarshal([]byte(args), &data); err == nil && args != "" {
			toolCallID++
		}
	}
	defer streaming.CallWithDone(ctx, payload.StreamingFunc) //nolint:errcheck

	for streamResponse := range responseChan {
		chunk := streamResponse.Result
		response.Result += streamResponse.Result
		response.IsTruncated = streamResponse.IsTruncated
		if streamResponse.FunctionCall != nil {
			response.FunctionCall = streamResponse.FunctionCall
			functionCall := streamResponse.FunctionCall
			toolCall := streaming.ToolCall{
				ID:        fmt.Sprintf("call-%d", toolCallID),
				Name:      functionCall.Name,
				Arguments: functionCall.Arguments,
			}
			incToolCallID(functionCall.Arguments)
			if err := streaming.CallWithToolCall(ctx, payload.StreamingFunc, toolCall); err != nil {
				return nil, fmt.Errorf("streaming func returned an error: %w", err)
			}
			if err := streaming.CallWithReasoning(ctx, payload.StreamingFunc, functionCall.Thoughts); err != nil {
				return nil, fmt.Errorf("streaming func returned an error: %w", err)
			}
		}

		if err := streaming.CallWithText(ctx, payload.StreamingFunc, chunk); err != nil {
			return nil, fmt.Errorf("streaming func returned an error: %w", err)
		}

		if streamResponse.IsEnd {
			break
		}
	}
	return &response, nil
}

type errorMessage struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}
