package azureaisearch_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/sayerxofficial/langchaingo/chains"
	"github.com/sayerxofficial/langchaingo/embeddings"
	"github.com/sayerxofficial/langchaingo/llms/openai"
	"github.com/sayerxofficial/langchaingo/schema"
	"github.com/sayerxofficial/langchaingo/vectorstores"
	"github.com/sayerxofficial/langchaingo/vectorstores/azureaisearch"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func checkEnvVariables(t *testing.T) {
	t.Helper()

	azureaisearchEndpoint := os.Getenv(azureaisearch.EnvironmentVariableEndpoint)
	if azureaisearchEndpoint == "" {
		t.Skipf("Must set %s to run test", azureaisearch.EnvironmentVariableEndpoint)
	}

	azureaisearchAPIKey := os.Getenv(azureaisearch.EnvironmentVariableAPIKey)
	if azureaisearchAPIKey == "" {
		t.Skipf("Must set %s to run test", azureaisearch.EnvironmentVariableAPIKey)
	}

	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
	}
}

func setIndex(t *testing.T, storer azureaisearch.Store, indexName string) {
	t.Helper()

	if err := storer.CreateIndex(t.Context(), indexName); err != nil {
		t.Fatalf("error creating index: %v\n", err)
	}
}

func removeIndex(t *testing.T, storer azureaisearch.Store, indexName string) {
	t.Helper()

	if err := storer.DeleteIndex(t.Context(), indexName); err != nil {
		t.Fatalf("error deleting index: %v\n", err)
	}
}

func setLLM(t *testing.T) *openai.LLM {
	t.Helper()

	openaiOpts := []openai.Option{}

	if openAIBaseURL := os.Getenv("OPENAI_BASE_URL"); openAIBaseURL != "" {
		openaiOpts = append(openaiOpts,
			openai.WithBaseURL(openAIBaseURL),
			openai.WithAPIType(openai.APITypeAzure),
			openai.WithEmbeddingModel("text-embedding-ada-002"),
			openai.WithModel("gpt-4"),
		)
	}

	llm, err := openai.New(openaiOpts...)
	if err != nil {
		t.Fatalf("error setting openAI embedded: %v\n", err)
	}

	return llm
}

func TestAzureaiSearchStoreRest(t *testing.T) {
	t.Parallel()

	checkEnvVariables(t)
	indexName := uuid.New().String()

	llm := setLLM(t)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	storer, err := azureaisearch.New(
		azureaisearch.WithEmbedder(e),
	)
	require.NoError(t, err)

	setIndex(t, storer, indexName)
	defer removeIndex(t, storer, indexName)

	_, err = storer.AddDocuments(t.Context(), []schema.Document{
		{PageContent: "tokyo"},
		{PageContent: "potato"},
	}, vectorstores.WithNameSpace(indexName))
	require.NoError(t, err)

	docs, err := storer.SimilaritySearch(t.Context(), "japan", 1, vectorstores.WithNameSpace(indexName))
	require.NoError(t, err)
	require.Len(t, docs, 1)
	require.Equal(t, "tokyo", docs[0].PageContent)
}

func TestAzureaiSearchStoreRestWithScoreThreshold(t *testing.T) {
	t.Parallel()

	checkEnvVariables(t)
	indexName := uuid.New().String()

	llm := setLLM(t)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	storer, err := azureaisearch.New(
		azureaisearch.WithEmbedder(e),
	)
	require.NoError(t, err)

	setIndex(t, storer, indexName)
	defer removeIndex(t, storer, indexName)

	_, err = storer.AddDocuments(t.Context(), []schema.Document{
		{PageContent: "Tokyo"},
		{PageContent: "Yokohama"},
		{PageContent: "Osaka"},
		{PageContent: "Nagoya"},
		{PageContent: "Sapporo"},
		{PageContent: "Fukuoka"},
		{PageContent: "Dublin"},
		{PageContent: "Paris"},
		{PageContent: "London "},
		{PageContent: "New York"},
	}, vectorstores.WithNameSpace(indexName))
	require.NoError(t, err)
	time.Sleep(time.Second)
	// test with a score threshold of 0.84, expected 6 documents
	docs, err := storer.SimilaritySearch(t.Context(),
		"Which of these are cities in Japan", 10,
		vectorstores.WithScoreThreshold(0.84),
		vectorstores.WithNameSpace(indexName))
	require.NoError(t, err)
	require.Len(t, docs, 6)
}

func TestAzureaiSearchAsRetriever(t *testing.T) {
	t.Parallel()

	checkEnvVariables(t)
	indexName := uuid.New().String()

	llm := setLLM(t)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	storer, err := azureaisearch.New(
		azureaisearch.WithEmbedder(e),
	)
	require.NoError(t, err)

	setIndex(t, storer, indexName)
	defer removeIndex(t, storer, indexName)

	_, err = storer.AddDocuments(
		t.Context(),
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
		},
		vectorstores.WithNameSpace(indexName),
	)
	require.NoError(t, err)

	time.Sleep(time.Second)

	result, err := chains.Run(
		t.Context(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(&storer, 1, vectorstores.WithNameSpace(indexName)),
		),
		"What color is the desk?",
	)
	require.NoError(t, err)
	require.Contains(t, strings.ToLower(result), "orange", "expected orange in result")
}

func TestAzureaiSearchAsRetrieverWithScoreThreshold(t *testing.T) {
	t.Parallel()
	checkEnvVariables(t)
	indexName := uuid.New().String()

	llm := setLLM(t)
	e, err := embeddings.NewEmbedder(llm)
	require.NoError(t, err)

	storer, err := azureaisearch.New(
		azureaisearch.WithEmbedder(e),
	)
	require.NoError(t, err)

	setIndex(t, storer, indexName)
	defer removeIndex(t, storer, indexName)

	_, err = storer.AddDocuments(
		t.Context(),
		[]schema.Document{
			{PageContent: "The color of the house is blue."},
			{PageContent: "The color of the car is red."},
			{PageContent: "The color of the desk is orange."},
			{PageContent: "The color of the lamp beside the desk is black."},
			{PageContent: "The color of the chair beside the desk is beige."},
		},
		vectorstores.WithNameSpace(indexName),
	)
	require.NoError(t, err)
	time.Sleep(time.Second)
	result, err := chains.Run(
		t.Context(),
		chains.NewRetrievalQAFromLLM(
			llm,
			vectorstores.ToRetriever(&storer, 5,
				vectorstores.WithNameSpace(indexName),
				vectorstores.WithScoreThreshold(0.8)),
		),
		"What colors is each piece of furniture next to the desk?",
	)
	require.NoError(t, err)

	require.Contains(t, strings.ToLower(result), "black", "expected black in result")
	require.Contains(t, strings.ToLower(result), "beige", "expected beige in result")
}
