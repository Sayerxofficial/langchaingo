package chains

import (
	"net/http"
	"os"
	"testing"

	"github.com/sayerxofficial/langchaingo/documentloaders"
	"github.com/sayerxofficial/langchaingo/internal/httprr"
	"github.com/sayerxofficial/langchaingo/llms/openai"
	"github.com/sayerxofficial/langchaingo/schema"
	"github.com/sayerxofficial/langchaingo/textsplitter"

	"github.com/stretchr/testify/require"
)

func loadTestData(t *testing.T) []schema.Document {
	t.Helper()
	ctx := t.Context()

	file, err := os.Open("./testdata/mouse_story.txt")
	require.NoError(t, err)

	docs, err := documentloaders.NewText(file).LoadAndSplit(
		ctx,
		textsplitter.NewRecursiveCharacter(),
	)
	require.NoError(t, err)

	return docs
}

// createOpenAILLMForTest creates an OpenAI LLM with httprr support for testing.
func createOpenAILLMForTest(t *testing.T) *openai.LLM {
	t.Helper()
	httprr.SkipIfNoCredentialsAndRecordingMissing(t, "OPENAI_API_KEY")

	rr := httprr.OpenForTest(t, http.DefaultTransport)

	// Only run tests in parallel when not recording
	if !rr.Recording() {
		t.Parallel()
	}

	opts := []openai.Option{
		openai.WithHTTPClient(rr.Client()),
	}

	// Only add fake token when NOT recording (i.e., during replay)
	if !rr.Recording() {
		opts = append(opts, openai.WithToken("test-api-key"))
	}
	// When recording, openai.New() will read OPENAI_API_KEY from environment

	llm, err := openai.New(opts...)
	require.NoError(t, err)
	return llm
}

func TestStuffSummarization(t *testing.T) {
	ctx := t.Context()

	llm := createOpenAILLMForTest(t)

	docs := loadTestData(t)

	chain := LoadStuffSummarization(llm)
	_, err := Call(
		ctx,
		chain,
		map[string]any{"input_documents": docs},
	)
	require.NoError(t, err)
}

func TestRefineSummarization(t *testing.T) {
	ctx := t.Context()

	llm := createOpenAILLMForTest(t)

	docs := loadTestData(t)

	chain := LoadRefineSummarization(llm)
	_, err := Call(
		ctx,
		chain,
		map[string]any{"input_documents": docs},
	)
	require.NoError(t, err)
}

func TestMapReduceSummarization(t *testing.T) {
	ctx := t.Context()

	llm := createOpenAILLMForTest(t)

	docs := loadTestData(t)

	chain := LoadMapReduceSummarization(llm)
	_, err := Run(
		ctx,
		chain,
		docs,
	)
	require.NoError(t, err)
}
