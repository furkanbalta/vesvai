package agent

import (
	"context"

	"github.com/vesvai/vesvai/internal/core/hook"
	"github.com/vesvai/vesvai/internal/llm"
)

type MessageInput struct {
	Text  string
	Calls []llm.ToolCall
}

var messageInputHook = hook.NewHook[MessageInput]()

func OnMessageInput(fn func(MessageInput) MessageInput) {
	messageInputHook.Add(fn)
}

func expandInput(input string) MessageInput {
	return messageInputHook.Apply(MessageInput{Text: input})
}

type ModelResolve struct {
	Ctx   context.Context
	Agent *Agent
	Input string
}

var ModelResolveHook = hook.NewHook[ModelResolve]()

func OnModelResolve(fn func(ModelResolve) ModelResolve) {
	ModelResolveHook.Add(fn)
}

func resolveModel(ctx context.Context, a *Agent, input string) {
	ModelResolveHook.Apply(ModelResolve{Ctx: ctx, Agent: a, Input: input})
}
