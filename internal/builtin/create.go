package builtin

import (
	"github.com/vesvai/vesvai/internal/builtin/agents"
	"github.com/vesvai/vesvai/internal/builtin/middlewares"
	"github.com/vesvai/vesvai/internal/builtin/reminders"
	"github.com/vesvai/vesvai/internal/builtin/tools"
	"github.com/vesvai/vesvai/internal/core/config"
	"github.com/vesvai/vesvai/internal/core/event"
	"github.com/vesvai/vesvai/internal/decision"
	"github.com/vesvai/vesvai/internal/llm"
	"github.com/vesvai/vesvai/internal/memory"
	"github.com/vesvai/vesvai/internal/session"
	"github.com/vesvai/vesvai/internal/vfs"
)

type Options struct {
	LLM      *llm.Manager
	Decision *decision.Manager
	Config   *config.Config
	Bus      event.Bus
	Memory   *memory.Manager
}

func Create(fs *vfs.VFS, sess *session.Manager, opts Options) error {
	agents.Create(fs)
	middlewares.Create(fs, middlewares.Deps{
		Config:   opts.Config,
		LLM:      opts.LLM,
		Decision: opts.Decision,
		Sessions: sess,
	})
	tools.Create(fs, sess)
	if opts.Memory != nil {
		if err := opts.Memory.Start(); err != nil {
			return err
		}
		if err := memory.RegisterTools(opts.Memory); err != nil {
			return err
		}
	}
	if err := reminders.Create(opts.Bus); err != nil {
		return err
	}
	return nil
}
