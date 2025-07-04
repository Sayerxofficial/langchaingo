package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/sayerxofficial/langchaingo/internal/httprr"
	"github.com/sayerxofficial/langchaingo/llms"
	"github.com/sayerxofficial/langchaingo/llms/streaming"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestOpenAIClient(t *testing.T, opts ...Option) *LLM {
	t.Helper()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Configure OpenAI client based on recording vs replay mode
	clientOpts := []Option{WithHTTPClient(rr.Client())}

	// Only add fake token when NOT recording (i.e., during replay)
	if !rr.Recording() {
		clientOpts = append(clientOpts, WithToken("fake-api-key-for-testing"))
	} else {
		clientOpts = append(clientOpts, WithToken(os.Getenv("OPENAI_API_KEY")))
	}

	// Add any additional options passed to the function
	clientOpts = append(clientOpts, opts...)

	t.Logf("Creating OpenAI client with recording=%v", rr.Recording())
	llm, err := New(clientOpts...)
	require.NoError(t, err)
	return llm
}

func newTestDeepSeekClient(t *testing.T, opts ...Option) *LLM {
	t.Helper()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "DEEPSEEK_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Configure OpenAI client based on recording vs replay mode
	clientOpts := []Option{
		WithBaseURL("https://api.deepseek.com"),
		WithHTTPClient(rr.Client()),
	}

	// Only add fake token when NOT recording (i.e., during replay)
	if !rr.Recording() {
		clientOpts = append(clientOpts, WithToken("fake-api-key-for-testing"))
	} else {
		clientOpts = append(clientOpts, WithToken(os.Getenv("DEEPSEEK_API_KEY")))
	}

	// Add any additional options passed to the function
	clientOpts = append(clientOpts, opts...)

	t.Logf("Creating DeepSeek client with recording=%v", rr.Recording())
	llm, err := New(clientOpts...)
	require.NoError(t, err)
	return llm
}

func newTestOpenRouterClient(t *testing.T, opts ...Option) *LLM {
	t.Helper()

	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENROUTER_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Configure OpenAI client based on recording vs replay mode
	clientOpts := []Option{
		WithBaseURL("https://openrouter.ai/api/v1"),
		WithHTTPClient(rr.Client()),
	}

	// Only add fake token when NOT recording (i.e., during replay)
	if !rr.Recording() {
		clientOpts = append(clientOpts, WithToken("fake-api-key-for-testing"))
	} else {
		clientOpts = append(clientOpts, WithToken(os.Getenv("OPENROUTER_API_KEY")))
	}

	// Add any additional options passed to the function
	clientOpts = append(clientOpts, opts...)

	t.Logf("Creating OpenRouter client with recording=%v", rr.Recording())
	llm, err := New(clientOpts...)
	require.NoError(t, err)
	return llm
}

type testEnv struct {
	name string
	init func(t *testing.T, opts ...Option) *LLM
	opts []Option

	// reasoning options
	ropt llms.CallOption
	rout bool
}

func getCompletionTests() []testEnv {
	var openRouterModels = []string{ //nolint:gofumpt
		"anthropic/claude-3.7-sonnet:thinking",
		"anthropic/claude-3.7-sonnet",
		"deepseek/deepseek-chat",
		"deepseek/deepseek-r1",
		"google/gemini-2.0-flash-001",
		"google/gemini-2.0-flash-lite-001",
		"google/gemini-2.5-flash-preview:thinking",
		"mistralai/mistral-medium-3",
		"mistralai/mistral-nemo",
		"openai/gpt-4.1-mini",
		"openai/gpt-4.1",
		"openai/gpt-4o-mini",
		"openai/gpt-4o",
		"openai/o3-mini-high",
		"openai/o3-mini",
		"openai/o4-mini-high",
		"openai/o4-mini",
		"qwen/qwen-turbo",
		"qwen/qwen3-235b-a22b",
	}
	tests := []testEnv{
		{
			name: "openai",
			init: newTestOpenAIClient,
			opts: []Option{WithModel("gpt-4.1-mini")},
		},
		{
			name: "deepseek",
			init: newTestDeepSeekClient,
			opts: []Option{WithModel("deepseek-reasoner")},
		},
	}
	for _, model := range openRouterModels {
		tests = append(tests, testEnv{
			name: "openrouter-" + strings.ReplaceAll(strings.Split(model, "/")[1], ":", "-"),
			init: newTestOpenRouterClient,
			opts: []Option{WithModel(model)},
		})
	}
	return tests
}

func getReasoningTests() []testEnv {
	var openRouterModels = []string{ //nolint:gofumpt
		"anthropic/claude-3.7-sonnet:thinking",
		"deepseek/deepseek-r1",
		"google/gemini-2.5-flash-preview:thinking",
		"openai/o3-mini-high",
		"openai/o3-mini",
		"openai/o4-mini-high",
		"openai/o4-mini",
	}
	reasoningOption := llms.WithReasoning(llms.ReasoningHigh, 0)
	tests := []testEnv{
		{
			name: "openai-o1",
			init: newTestOpenAIClient,
			opts: []Option{WithModel("o1")},
			ropt: reasoningOption,
			rout: false,
		},
		{
			name: "openai-o3",
			init: newTestOpenAIClient,
			opts: []Option{WithModel("o3")},
			ropt: reasoningOption,
			rout: false,
		},
		{
			name: "openai-o3-mini",
			init: newTestOpenAIClient,
			opts: []Option{WithModel("o3-mini")},
			ropt: reasoningOption,
			rout: false,
		},
		{
			name: "openai-o4-mini",
			init: newTestOpenAIClient,
			opts: []Option{WithModel("o4-mini")},
			ropt: reasoningOption,
			rout: false,
		},
		{
			name: "deepseek",
			init: newTestDeepSeekClient,
			opts: []Option{WithModel("deepseek-reasoner")},
			rout: true,
		},
	}
	clientReasoningOptions := []Option{
		WithUsingReasoningMaxTokens(),
		WithModernReasoningFormat(),
	}
	for _, model := range openRouterModels {
		// OpenAI doesn't provide reasoning content except for o4-mini and o4-mini-high
		expectResoningContent := !strings.Contains(model, "openai") ||
			strings.Contains(model, "o4-mini") || strings.Contains(model, "o4-mini-high")
		normModelName := strings.ReplaceAll(strings.Split(model, "/")[1], ":", "-")
		tests = append(tests, testEnv{
			name: "openrouter-" + normModelName + "-effort",
			init: newTestOpenRouterClient,
			opts: append([]Option{WithModel(model)}, clientReasoningOptions...),
			ropt: reasoningOption,
			rout: expectResoningContent,
		})
		if strings.HasSuffix(model, ":thinking") {
			tests = append(tests, testEnv{
				name: "openrouter-" + normModelName + "-tokens",
				init: newTestOpenRouterClient,
				opts: append([]Option{WithModel(model)}, clientReasoningOptions...),
				ropt: llms.WithReasoning(llms.ReasoningNone, 2048),
				rout: expectResoningContent,
			})
		}
	}
	return tests
}

func getToolCallTests(multiToolCalls bool) []testEnv {
	var openRouterModels = []string{ //nolint:gofumpt
		"deepseek/deepseek-chat",
		"google/gemini-2.0-flash-001",
		"google/gemini-2.5-flash-preview:thinking",
		"mistralai/mistral-medium-3",
		"openai/gpt-4.1-mini",
		"openai/gpt-4.1",
		"openai/gpt-4o-mini",
		"openai/gpt-4o",
	}
	if !multiToolCalls {
		openRouterModels = append(openRouterModels,
			"anthropic/claude-3.7-sonnet:thinking",
			"anthropic/claude-3.7-sonnet",
			"openai/o3-mini-high",
			"openai/o3-mini",
			"openai/o4-mini-high",
			"openai/o4-mini",
			"qwen/qwen-turbo",
		)
	}
	tests := []testEnv{
		{
			name: "openai",
			init: newTestOpenAIClient,
			opts: []Option{WithModel("gpt-4.1")},
		},
		{
			name: "deepseek",
			init: newTestDeepSeekClient,
			opts: []Option{WithModel("deepseek-chat")},
		},
	}
	for _, model := range openRouterModels {
		tests = append(tests, testEnv{
			name: "openrouter-" + strings.ReplaceAll(strings.Split(model, "/")[1], ":", "-"),
			init: newTestOpenRouterClient,
			opts: []Option{WithModel(model)},
		})
	}
	return tests
}

func TestMultiContentText(t *testing.T) {
	t.Parallel()

	parts := []llms.ContentPart{
		llms.TextPart("I'm a pomeranian"),
		llms.TextPart("What kind of mammal am I?"),
	}
	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	tests := getCompletionTests()
	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			llm := test.init(t, test.opts...)

			resp, err := llm.GenerateContent(t.Context(), messages)
			require.NoError(t, err)

			assert.NotEmpty(t, resp.Choices)
			c1 := resp.Choices[0]
			assert.Regexp(t, "dog|canid", strings.ToLower(c1.Content))
		})
	}
}

func TestMultiContentTextWithReasoning(t *testing.T) {
	t.Parallel()

	parts := []llms.ContentPart{
		llms.TextPart("What is the factorial of 5? Show your work. Think before you answer."),
	}
	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	testFunc := func(t *testing.T, test testEnv, isStreaming bool) {
		t.Helper()

		var (
			content, reasoningContent strings.Builder
			streamDone                bool
		)
		opts := []llms.CallOption{
			llms.WithMaxTokens(8192), // enough max tokens for reasoning
		}

		if isStreaming {
			opts = append(opts, llms.WithStreamingFunc(func(_ context.Context, chunk streaming.Chunk) error {
				switch chunk.Type {
				case streaming.ChunkTypeNone:
					// skip none chunks
				case streaming.ChunkTypeText:
					content.WriteString(chunk.Content)
				case streaming.ChunkTypeReasoning:
					reasoningContent.WriteString(chunk.ReasoningContent)
				case streaming.ChunkTypeToolCall:
					// skip tool calls
				case streaming.ChunkTypeDone:
					streamDone = true
				}
				return nil
			}))
		}

		if test.ropt != nil {
			opts = append(opts, test.ropt)
		}

		llm := test.init(t, test.opts...)
		resp, err := llm.GenerateContent(t.Context(), messages, opts...)
		require.NoError(t, err)

		assert.NotEmpty(t, resp.Choices)
		c1 := resp.Choices[0]
		assert.Contains(t, strings.ToLower(c1.Content), "120")

		if test.rout {
			assert.NotEmpty(t, c1.ReasoningContent)
		} else {
			assert.Empty(t, c1.ReasoningContent)
		}

		if isStreaming {
			assert.True(t, streamDone)
			assert.Equal(t, content.String(), c1.Content)
			assert.Equal(t, reasoningContent.String(), c1.ReasoningContent)
		}
	}

	tests := getReasoningTests()
	for idx := range tests {
		test := tests[idx]
		t.Run("synchronous-"+test.name, func(t *testing.T) {
			t.Parallel()

			testFunc(t, test, false)
		})
		t.Run("streaming-"+test.name, func(t *testing.T) {
			t.Parallel()

			testFunc(t, test, true)
		})
	}
}

func TestMultiContentTextChatSequence(t *testing.T) {
	t.Parallel()

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Name some countries")},
		},
		{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextPart("Spain and Lesotho")},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Which if these is larger?")},
		},
	}

	tests := getCompletionTests()
	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			llm := test.init(t, test.opts...)

			resp, err := llm.GenerateContent(t.Context(), messages)
			require.NoError(t, err)

			assert.NotEmpty(t, resp.Choices)
			c1 := resp.Choices[0]
			assert.Regexp(t, "spain.*larger", strings.ToLower(c1.Content))
		})
	}
}

func TestMultiContentImage(t *testing.T) {
	t.Parallel()

	llm := newTestOpenAIClient(t, WithModel("gpt-4o"))

	parts := []llms.ContentPart{
		llms.ImageURLPart("https://github.com/sayerxofficial/langchaingo/blob/main/docs/static/img/parrot-icon.png?raw=true"), //nolint:lll
		llms.TextPart("describe this image in detail"),
	}
	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	resp, err := llm.GenerateContent(t.Context(), messages, llms.WithMaxTokens(300))
	require.NoError(t, err)

	assert.NotEmpty(t, resp.Choices)
	c1 := resp.Choices[0]
	assert.Contains(t, strings.ToLower(c1.Content), "parrot")
}

func TestWithStreaming(t *testing.T) {
	t.Parallel()

	parts := []llms.ContentPart{
		llms.TextPart("I'm a pomeranian"),
		llms.TextPart("Tell me more about my taxonomy"),
	}
	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	tests := getCompletionTests()
	for idx := range tests {
		test := tests[idx]
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			llm := test.init(t, test.opts...)

			var (
				text       strings.Builder
				reasoning  strings.Builder
				streamDone bool
			)
			resp, err := llm.GenerateContent(t.Context(), messages,
				llms.WithStreamingFunc(func(_ context.Context, chunk streaming.Chunk) error {
					switch chunk.Type {
					case streaming.ChunkTypeNone:
						// skip none chunks
					case streaming.ChunkTypeText:
						text.WriteString(chunk.Content)
					case streaming.ChunkTypeReasoning:
						reasoning.WriteString(chunk.ReasoningContent)
					case streaming.ChunkTypeToolCall:
						// skip tool calls
					case streaming.ChunkTypeDone:
						streamDone = true
					}
					return nil
				}),
			)
			require.NoError(t, err)

			assert.True(t, streamDone)
			assert.NotEmpty(t, resp.Choices)
			c1 := resp.Choices[0]
			assert.Regexp(t, "dog|canid", strings.ToLower(c1.Content))
			assert.Equal(t, text.String(), c1.Content)
			assert.Equal(t, reasoning.String(), c1.ReasoningContent)
		})
	}
}

//nolint:funlen
func TestFunctionCall(t *testing.T) {
	t.Parallel()

	parts := []llms.ContentPart{
		llms.TextPart("What is the weather like in Boston, MA?"),
	}
	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	functions := []llms.FunctionDefinition{
		{
			Name:        "getCurrentWeather",
			Description: "Get the current weather in a given location",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"location": {"type": "string", "description": "The city and state, e.g. San Francisco, CA"},
					"unit": {"type": "string", "enum": ["celsius", "fahrenheit"]}
				},
				"required": ["location"]
			}`),
		},
	}

	tests := getToolCallTests(false)

	for idx := range tests {
		test := tests[idx]
		t.Run("synchronous-"+test.name, func(t *testing.T) {
			t.Parallel()

			llm := test.init(t, test.opts...)

			resp, err := llm.GenerateContent(t.Context(), messages, llms.WithFunctions(functions))
			require.NoError(t, err)

			assert.NotEmpty(t, resp.Choices)
			c1 := resp.Choices[0]
			if c1.StopReason != "tool_calls" {
				t.Logf("Unexpected stop reason (expected tool_calls): %s", c1.StopReason)
			}
			assert.NotNil(t, c1.FuncCall)
			assert.Len(t, c1.ToolCalls, 1)

			if len(c1.ToolCalls) >= 1 && c1.FuncCall != nil {
				assert.Equal(t, c1.ToolCalls[0].FunctionCall.Name, c1.FuncCall.Name)
				assert.Equal(t, c1.ToolCalls[0].FunctionCall.Arguments, c1.FuncCall.Arguments)
				assert.Equal(t, "getCurrentWeather", c1.FuncCall.Name)
			}
		})
	}

	for idx := range tests {
		test := tests[idx]
		t.Run("streaming-"+test.name, func(t *testing.T) {
			t.Parallel()

			llm := test.init(t, test.opts...)

			var (
				toolCall   streaming.ToolCall
				streamDone bool
			)
			streamingFunc := func(_ context.Context, chunk streaming.Chunk) error {
				switch chunk.Type {
				case streaming.ChunkTypeNone:
					// skip none chunks
				case streaming.ChunkTypeText:
					// skip text chunks
				case streaming.ChunkTypeReasoning:
					// skip reasoning chunks
				case streaming.ChunkTypeToolCall:
					toolCall.ID = chunk.ToolCall.ID
					toolCall.Name = chunk.ToolCall.Name
					toolCall.Arguments += chunk.ToolCall.Arguments
				case streaming.ChunkTypeDone:
					streamDone = true
				}
				return nil
			}

			resp, err := llm.GenerateContent(
				t.Context(),
				messages,
				llms.WithFunctions(functions),
				llms.WithStreamingFunc(streamingFunc),
			)
			require.NoError(t, err)

			assert.True(t, streamDone)
			assert.NotEmpty(t, resp.Choices)
			c1 := resp.Choices[0]
			if c1.StopReason != "tool_calls" {
				t.Logf("Unexpected stop reason (expected tool_calls): %s", c1.StopReason)
			}
			assert.NotNil(t, c1.FuncCall)
			assert.Len(t, c1.ToolCalls, 1)

			if len(c1.ToolCalls) >= 1 && c1.FuncCall != nil {
				assert.Equal(t, c1.ToolCalls[0].FunctionCall.Name, toolCall.Name)
				assert.Equal(t, c1.ToolCalls[0].FunctionCall.Arguments, toolCall.Arguments)
				assert.Equal(t, c1.ToolCalls[0].FunctionCall.Name, c1.FuncCall.Name)
				assert.Equal(t, c1.ToolCalls[0].FunctionCall.Arguments, c1.FuncCall.Arguments)
				assert.Equal(t, "getCurrentWeather", toolCall.Name)
			}
		})
	}
}

//nolint:funlen,cyclop
func TestFunctionParallelCall(t *testing.T) {
	t.Parallel()

	parts := []llms.ContentPart{
		llms.TextPart("What are the weather and time in Boston, MA?"),
	}
	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	functions := []llms.FunctionDefinition{
		{
			Name:        "getCurrentWeather",
			Description: "Get the current weather in a given location",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"location": {"type": "string", "description": "The city and state, e.g. San Francisco, CA"},
					"unit": {"type": "string", "enum": ["celsius", "fahrenheit"]}
				},
				"required": ["location"]
			}`),
		},
		{
			Name:        "getCurrentTime",
			Description: "Get the current time in a given location",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"location": {"type": "string", "description": "The city and state, e.g. San Francisco, CA"}
				},
				"required": ["location"]
			}`),
		},
	}

	tests := getToolCallTests(true)

	for idx := range tests {
		test := tests[idx]
		t.Run("synchronous-"+test.name, func(t *testing.T) {
			t.Parallel()

			llm := test.init(t, test.opts...)

			resp, err := llm.GenerateContent(t.Context(), messages, llms.WithFunctions(functions))
			require.NoError(t, err)

			assert.NotEmpty(t, resp.Choices)
			c1 := resp.Choices[0]
			if c1.StopReason != "tool_calls" {
				t.Logf("Unexpected stop reason (expected tool_calls): %s", c1.StopReason)
			}
			assert.NotNil(t, c1.FuncCall)
			assert.Len(t, c1.ToolCalls, 2)

			if len(c1.ToolCalls) >= 2 && c1.FuncCall != nil {
				// First tool call is about weather and it keeps in FuncCall
				assert.NotNil(t, c1.ToolCalls[0].FunctionCall)
				assert.Equal(t, c1.ToolCalls[0].FunctionCall.Name, c1.FuncCall.Name)
				assert.Equal(t, c1.ToolCalls[0].FunctionCall.Arguments, c1.FuncCall.Arguments)
				assert.Equal(t, "getCurrentWeather", c1.FuncCall.Name)
				// Second tool call is about time
				assert.NotNil(t, c1.ToolCalls[1].FunctionCall)
				assert.Equal(t, c1.ToolCalls[1].FunctionCall.Name, "getCurrentTime")
			}
		})
	}

	for idx := range tests {
		test := tests[idx]
		t.Run("streaming-"+test.name, func(t *testing.T) {
			t.Parallel()

			llm := test.init(t, test.opts...)

			var streamDone bool
			toolCalls := make(map[string]*streaming.ToolCall)
			streamingFunc := func(_ context.Context, chunk streaming.Chunk) error {
				switch chunk.Type {
				case streaming.ChunkTypeNone:
					// skip none chunks
				case streaming.ChunkTypeText:
					// skip text chunks
				case streaming.ChunkTypeReasoning:
					// skip reasoning chunks
				case streaming.ChunkTypeToolCall:
					toolCall, ok := toolCalls[chunk.ToolCall.ID]
					if !ok {
						toolCall = &streaming.ToolCall{}
						toolCalls[chunk.ToolCall.ID] = toolCall
					}
					toolCall.ID = chunk.ToolCall.ID
					toolCall.Name = chunk.ToolCall.Name
					toolCall.Arguments += chunk.ToolCall.Arguments
				case streaming.ChunkTypeDone:
					streamDone = true
				}
				return nil
			}

			resp, err := llm.GenerateContent(
				t.Context(),
				messages,
				llms.WithFunctions(functions),
				llms.WithStreamingFunc(streamingFunc),
			)
			require.NoError(t, err)

			assert.True(t, streamDone)
			assert.NotEmpty(t, resp.Choices)
			c1 := resp.Choices[0]
			if c1.StopReason != "tool_calls" {
				t.Logf("Unexpected stop reason (expected tool_calls): %s", c1.StopReason)
			}
			assert.NotNil(t, c1.FuncCall)
			assert.Len(t, c1.ToolCalls, 2)

			if len(c1.ToolCalls) >= 2 && c1.FuncCall != nil {
				// First tool call is about weather and it keeps in FuncCall
				assert.NotNil(t, c1.ToolCalls[0].FunctionCall)
				assert.Equal(t, c1.ToolCalls[0].FunctionCall.Name, c1.FuncCall.Name)
				assert.Equal(t, c1.ToolCalls[0].FunctionCall.Arguments, c1.FuncCall.Arguments)
				assert.Equal(t, "getCurrentWeather", c1.FuncCall.Name)
				// Second tool call is about time
				assert.NotNil(t, c1.ToolCalls[1].FunctionCall)
				assert.Equal(t, c1.ToolCalls[1].FunctionCall.Name, "getCurrentTime")

				// Check that the tool calls are the same as the ones in the streaming function
				assert.Len(t, toolCalls, 2)
				for _, tc := range c1.ToolCalls {
					toolCall, ok := toolCalls[tc.ID]
					assert.True(t, ok)
					assert.Equal(t, tc.ID, toolCall.ID)
					assert.Equal(t, tc.FunctionCall.Name, toolCall.Name)
					assert.Equal(t, tc.FunctionCall.Arguments, toolCall.Arguments)
				}
			}
		})
	}
}
