package agents

import (
	"regexp"
	"testing"

	"github.com/sayerxofficial/langchaingo/chains"
	"github.com/sayerxofficial/langchaingo/httputil"
	"github.com/sayerxofficial/langchaingo/internal/httprr"
	"github.com/sayerxofficial/langchaingo/llms/openai"
	"github.com/sayerxofficial/langchaingo/memory"
	"github.com/sayerxofficial/langchaingo/tools"

	"github.com/stretchr/testify/require"
)

func TestConversationalWithMemory(t *testing.T) {
	t.Parallel()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, httputil.DefaultTransport)
	// Configure OpenAI client with httprr
	opts := []openai.Option{
		openai.WithModel("gpt-4o"),
		openai.WithHTTPClient(rr.Client()),
	}
	if rr.Replaying() {
		opts = append(opts, openai.WithToken("test-api-key"))
	}

	llm, err := openai.New(opts...)
	require.NoError(t, err)

	executor, err := Initialize(
		llm,
		[]tools.Tool{tools.Calculator{}},
		ConversationalReactDescription,
		WithMemory(memory.NewConversationBuffer()),
	)
	require.NoError(t, err)

	_, err = chains.Run(t.Context(), executor, "Hi! my name is Bob and the year I was born is 1987")
	require.NoError(t, err)

	res, err := chains.Run(t.Context(), executor, "What is the year I was born times 34")
	require.NoError(t, err)
	expectedRe := "67,?558"
	if !regexp.MustCompile(expectedRe).MatchString(res) {
		t.Errorf("result does not contain the crrect answer '67558', got: %s", res)
	}
}
