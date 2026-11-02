package app

import (
	"sync"

	"github.com/sfossl/gdk/command"
	"github.com/sfossl/gdk/middleware"
)

type ComponentRegistration struct {
	CustomID string
	Handler  middleware.Handler
}

var (
	regMu      sync.RWMutex
	registered []command.Command

	compMu   sync.RWMutex
	regComps []ComponentRegistration
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

func RegisterComponent(customID string, handler middleware.Handler) {
	compMu.Lock()
	defer compMu.Unlock()
	regComps = append(regComps, ComponentRegistration{CustomID: customID, Handler: handler})
}

func RegisteredComponents() []ComponentRegistration {
	compMu.RLock()
	defer compMu.RUnlock()
	out := make([]ComponentRegistration, len(regComps))
	copy(out, regComps)
	return out
}

func (a *App) MountRegistered() *App {
	a.Commands(RegisteredCommands()...)
	for _, c := range RegisteredComponents() {
		a.Component(c.CustomID, c.Handler)
	}
	return a
}
