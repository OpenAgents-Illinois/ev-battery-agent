package agent

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
	"github.com/tmc/langchaingo/llms/googleai/vertex"

	"ev-battery-agent/internal/jira"
)

// embedderFunc takes a batch of texts and returns their embeddings.
type embedderFunc func(ctx context.Context, texts []string) ([][]float32, error)

// Factory initializes shared resources (LLM, embeddings, vector stores) once at startup.
// Use Chat to run the agent for a given vehicle model and user message.
type Factory struct {
	llm    *vertex.Vertex
	embed  embedderFunc
	stores map[string]*vectorStore // "R1S" → store, "R1T" → store
	jira   *jira.Client
}

// jiraToolDef is the Jira ticket tool schema passed to the LLM on every request.
var jiraToolDef = llms.Tool{
	Type: "function",
	Function: &llms.FunctionDefinition{
		Name: "fileEngineeringTicket",
		Description: "Creates an engineering ticket in Jira when a battery defect is detected. " +
			"Checks for an existing open ticket for this VIN first to avoid duplicates. " +
			"Severity must be one of: CRITICAL, WARNING, or INFO.",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"vin":             map[string]any{"type": "string", "description": "Vehicle Identification Number (VIN)"},
				"defectType":      map[string]any{"type": "string", "description": "Type of defect detected"},
				"technicalReason": map[string]any{"type": "string", "description": "Brief technical explanation without special characters"},
				"severity":        map[string]any{"type": "string", "description": "Severity level", "enum": []string{"CRITICAL", "WARNING", "INFO"}},
			},
			"required": []string{"vin", "defectType", "technicalReason", "severity"},
		},
	},
}

// NewFactory connects to Vertex AI and loads (or creates) per-vehicle embedding stores.
func NewFactory(ctx context.Context, projectID, location string) (*Factory, error) {
	llm, err := vertex.New(ctx,
		googleai.WithCloudProject(projectID),
		googleai.WithCloudLocation(location),
		googleai.WithDefaultModel("gemini-2.0-flash-001"),
		googleai.WithDefaultEmbeddingModel("text-embedding-004"),
	)
	if err != nil {
		return nil, fmt.Errorf("create LLM: %w", err)
	}

	embedFn := func(ctx context.Context, texts []string) ([][]float32, error) {
		return llm.CreateEmbedding(ctx, texts)
	}

	stores := make(map[string]*vectorStore)
	for _, model := range []string{"R1S", "R1T"} {
		store, err := buildStore(ctx, embedFn, "docs/"+model, fmt.Sprintf("embeddings/go_%s.json", model))
		if err != nil {
			return nil, fmt.Errorf("build store for %s: %w", model, err)
		}
		stores[model] = store
	}

	return &Factory{llm: llm, embed: embedFn, stores: stores, jira: jira.NewClient()}, nil
}

// buildStore loads embeddings from cache if present, otherwise embeds the docs and saves the cache.
func buildStore(ctx context.Context, embedFn embedderFunc, docsDir, cachePath string) (*vectorStore, error) {
	if store, err := loadCache(cachePath); err == nil {
		fmt.Printf("Loading embedding cache: %s (%d chunks)\n", cachePath, len(store.entries))
		return store, nil
	}

	fmt.Printf("No cache for %s — embedding now (one-time)...\n", docsDir)
	chunks, err := loadAndChunkDocs(docsDir)
	if err != nil {
		return nil, fmt.Errorf("load docs: %w", err)
	}
	fmt.Printf("  Embedding %d battery-relevant chunks...\n", len(chunks))

	store := &vectorStore{}
	if err := store.addChunks(ctx, embedFn, chunks); err != nil {
		return nil, fmt.Errorf("embed chunks: %w", err)
	}
	if err := store.saveCache(cachePath); err != nil {
		fmt.Printf("  Warning: could not save embedding cache: %v\n", err)
	} else {
		fmt.Printf("  Embedding cache saved: %s\n", cachePath)
	}
	return store, nil
}
