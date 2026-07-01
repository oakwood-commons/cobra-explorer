package tree

// CommandSelectedMsg is sent when the user selects a runnable command.
type CommandSelectedMsg struct {
	Node *CommandNode
}

// CommandHighlightedMsg is sent when the cursor moves to a different node.
type CommandHighlightedMsg struct {
	Node *CommandNode
}
