package agent

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

// Chat retrieves RAG context for the given vehicle model and runs the agent loop.
func (f *Factory) Chat(ctx context.Context, vehicleModel, userMessage string) (string, error) {
	model := strings.ToUpper(vehicleModel)
	store, ok := f.stores[model]
	if !ok {
		store = f.stores["R1S"]
	}

	// Embed query and prepend top-5 relevant doc chunks (non-fatal if it fails)
	if queryVecs, err := f.embed(ctx, []string{userMessage}); err == nil && len(queryVecs) > 0 {
		if chunks := store.search(ctx, queryVecs[0], 5); len(chunks) > 0 {
			var sb strings.Builder
			sb.WriteString("\n\nRelevant documentation:\n")
			for i, chunk := range chunks {
				sb.WriteString(fmt.Sprintf("[%d] %s\n\n", i+1, chunk))
			}
			userMessage += sb.String()
		}
	}

	return f.runAgentLoop(ctx, userMessage)
}

// runAgentLoop calls the LLM with tool-calling enabled, executing any tool calls,
// and repeats until the model returns a final text response (max 5 iterations).
func (f *Factory) runAgentLoop(ctx context.Context, userMessage string) (string, error) {
	// Vertex AI doesn't support a separate system role — prepend the system
	// prompt to the first human message instead.
	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeHuman, systemPrompt+"\n\n"+userMessage),
	}

	for range 5 {
		resp, err := f.llm.GenerateContent(ctx, messages,
			llms.WithTools([]llms.Tool{jiraToolDef}),
			llms.WithTemperature(0),
		)
		if err != nil {
			return "", fmt.Errorf("generate content: %w", err)
		}
		if len(resp.Choices) == 0 {
			return "", fmt.Errorf("empty response from LLM")
		}

		choice := resp.Choices[0]

		if len(choice.ToolCalls) == 0 {
			return choice.Content, nil
		}

		// Append the assistant turn (may include partial text + tool calls)
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
		messages = append(messages, llms.MessageContent{Role: llms.ChatMessageTypeAI, Parts: assistantParts})

		// Execute each tool and append results
		for _, tc := range choice.ToolCalls {
			result := f.dispatchTool(tc.FunctionCall.Name, tc.FunctionCall.Arguments)
			messages = append(messages, llms.MessageContent{
				Role: llms.ChatMessageTypeTool,
				Parts: []llms.ContentPart{
					llms.ToolCallResponse{ToolCallID: tc.ID, Name: tc.FunctionCall.Name, Content: result},
				},
			})
		}
	}

	return "", fmt.Errorf("agent exceeded maximum iterations")
}

// dispatchTool routes a tool call by name to the appropriate implementation.
func (f *Factory) dispatchTool(name, argsJSON string) string {
	switch name {
	case "fileEngineeringTicket":
		return f.jira.FileTicket(argsJSON)
	default:
		return "ERROR: Unknown tool: " + name
	}
}
