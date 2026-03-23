package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"

	"ev-battery-agent/internal/jira"
)

const systemPrompt = `You are an EV Battery Specialist. Analyze battery telemetry and determine if safety thresholds are violated.
Always analyze the telemetry data provided, regardless of its format.
When calling the fileEngineeringTicket tool, use only plain text — no special characters, quotes, or newlines in the vin, defectType, or technicalReason arguments.
Keep technicalReason under 80 characters.
Choose severity based on risk: CRITICAL for immediate safety hazards (thermal runaway, fire risk), WARNING for out-of-range but non-immediate issues, INFO for anomalies that need monitoring.`

const interactivePromptTmpl = `A Rivian EV owner has reported the following battery issue. Analyze it, determine if any safety thresholds are violated, and if so file a Jira ticket using the fileEngineeringTicket tool with severity CRITICAL, WARNING, or INFO.

Report: %s`

// embedderFunc is a function that takes a batch of texts and returns their embeddings.
type embedderFunc func(ctx context.Context, texts []string) ([][]float32, error)

// Factory initializes shared resources (LLM, embeddings, vector stores) once and creates agents per request.
type Factory struct {
	llm    *googleai.GoogleAI
	embed  embedderFunc
	stores map[string]*vectorStore // "R1S" and "R1T"
	jira   *jira.Client
}

// NewFactory builds the Factory: connects to Vertex AI and loads (or creates) embedding stores.
func NewFactory(ctx context.Context, projectID, location string) (*Factory, error) {
	llm, err := googleai.New(ctx,
		googleai.WithCloudProject(projectID),
		googleai.WithCloudLocation(location),
		googleai.WithDefaultModel("gemini-2.0-flash-001"),
		googleai.WithDefaultEmbeddingModel("text-embedding-004"),
	)
	if err != nil {
		return nil, fmt.Errorf("create LLM: %w", err)
	}

	// Wrap the LLM's CreateEmbedding as our embedderFunc
	embedFn := func(ctx context.Context, texts []string) ([][]float32, error) {
		return llm.CreateEmbedding(ctx, texts)
	}

	stores := make(map[string]*vectorStore)
	for _, model := range []string{"R1S", "R1T"} {
		cachePath := fmt.Sprintf("embeddings/go_%s.json", model)
		store, err := buildStore(ctx, embedFn, "docs/"+model, cachePath)
		if err != nil {
			return nil, fmt.Errorf("build store for %s: %w", model, err)
		}
		stores[model] = store
	}

	return &Factory{
		llm:    llm,
		embed:  embedFn,
		stores: stores,
		jira:   jira.NewClient(),
	}, nil
}

// buildStore loads from cache if available, otherwise embeds docs and saves cache.
func buildStore(ctx context.Context, embedFn embedderFunc, docsDir, cachePath string) (*vectorStore, error) {
	// Try loading from cache first
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

// Chat sends a user message through the agent for the given vehicle model.
// It retrieves relevant RAG context from the vehicle's embedding store before calling the LLM.
func (f *Factory) Chat(ctx context.Context, vehicleModel, userMessage string) (string, error) {
	model := strings.ToUpper(vehicleModel)
	store, ok := f.stores[model]
	if !ok {
		store = f.stores["R1S"]
	}

	// Embed query and retrieve top-5 relevant chunks
	queryVecs, err := f.embed(ctx, []string{userMessage})
	if err == nil && len(queryVecs) > 0 {
		relevant := store.search(ctx, queryVecs[0], 5)
		if len(relevant) > 0 {
			var ctxBuilder strings.Builder
			ctxBuilder.WriteString("\n\nRelevant documentation:\n")
			for i, chunk := range relevant {
				ctxBuilder.WriteString(fmt.Sprintf("[%d] %s\n\n", i+1, chunk))
			}
			userMessage = userMessage + ctxBuilder.String()
		}
	}
	// RAG failure is non-fatal — proceed without context

	return f.runAgentLoop(ctx, userMessage)
}

// runAgentLoop runs the LLM with tool-calling in a loop until the model stops calling tools.
func (f *Factory) runAgentLoop(ctx context.Context, userMessage string) (string, error) {
	jiraTool := llms.Tool{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name: "fileEngineeringTicket",
			Description: "Creates an engineering ticket in Jira when a battery defect is detected. " +
				"Checks for an existing open ticket for this VIN first to avoid duplicates. " +
				"Severity must be one of: CRITICAL, WARNING, or INFO.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"vin": map[string]any{
						"type":        "string",
						"description": "Vehicle Identification Number (VIN)",
					},
					"defectType": map[string]any{
						"type":        "string",
						"description": "Type of defect detected",
					},
					"technicalReason": map[string]any{
						"type":        "string",
						"description": "Brief technical explanation without special characters",
					},
					"severity": map[string]any{
						"type":        "string",
						"description": "Severity level",
						"enum":        []string{"CRITICAL", "WARNING", "INFO"},
					},
				},
				"required": []string{"vin", "defectType", "technicalReason", "severity"},
			},
		},
	}

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, userMessage),
	}

	for range 5 { // max 5 iterations (avoids infinite loops)
		resp, err := f.llm.GenerateContent(ctx, messages,
			llms.WithTools([]llms.Tool{jiraTool}),
			llms.WithTemperature(0),
		)
		if err != nil {
			return "", fmt.Errorf("generate content: %w", err)
		}
		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("empty response from LLM")
		}

		choice := resp.Choices[0]

		// No tool calls → final answer
		if len(choice.ToolCalls) == 0 {
			return choice.Content, nil
		}

		// Build assistant message with tool calls
		var assistantParts []llms.ContentPart
		if choice.Content != "" {
			assistantParts = append(assistantParts, llms.TextPart(choice.Content))
		}
		for _, tc := range choice.ToolCalls {
			assistantParts = append(assistantParts, llms.ToolCall{
				ID:   tc.ID,
				Type: "function",
				FunctionCall: &llms.FunctionCall{
					Name:      tc.FunctionCall.Name,
					Arguments: tc.FunctionCall.Arguments,
				},
			})
		}
		messages = append(messages, llms.MessageContent{
			Role:  llms.ChatMessageTypeAI,
			Parts: assistantParts,
		})

		// Execute each tool call and add results
		for _, tc := range choice.ToolCalls {
			var result string
			switch tc.FunctionCall.Name {
			case "fileEngineeringTicket":
				result = f.jira.FileTicket(tc.FunctionCall.Arguments)
			default:
				result = "ERROR: Unknown tool: " + tc.FunctionCall.Name
			}

			messages = append(messages, llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{
						ToolCallID: tc.ID,
						Name:       tc.FunctionCall.Name,
						Content:    result,
					},
				},
			})
		}
	}

	return "", fmt.Errorf("agent exceeded maximum iterations")
}

// DetectModel detects R1S or R1T from free text. Defaults to R1S.
func DetectModel(text string) string {
	upper := strings.ToUpper(text)
	if strings.Contains(upper, "R1T") {
		return "R1T"
	}
	return "R1S"
}

// InteractivePrompt formats a plain-English battery report for the agent.
func InteractivePrompt(report string) string {
	return fmt.Sprintf(interactivePromptTmpl, report)
}
