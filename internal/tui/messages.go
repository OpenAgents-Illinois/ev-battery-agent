package tui

// agentResultMsg is sent when the agent finishes processing a user message.
type agentResultMsg struct {
	text    string
	vehicle string
}

// agentErrMsg is sent when the agent returns an error.
type agentErrMsg struct{ err error }
