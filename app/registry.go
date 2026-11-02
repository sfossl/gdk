package app

import (
	"sync"

	"github.com/sfossl/gdk/command"
)

var (
	regMu      sync.RWMutex
	registered []command.Command
)

func Register(cmd command.Command) {
	if cmd.Disabled {
		return
	}
	regMu.Lock()
	defer regMu.Unlock()
	registered = append(registered, cmd)
}

func RegisteredCommands() []command.Command {
	regMu.RLock()
	defer regMu.RUnlock()
	out := make([]command.Command, len(registered))
	copy(out, registered)
	return out
}

func (a *App) MountRegistered() *App {
	return a.Commands(RegisteredCommands()...)
}

