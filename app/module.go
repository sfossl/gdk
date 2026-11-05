package app

import (
	"errors"
)

var (
	ErrNilModule = errors.New("module cannot be nil")
)

type Module interface {
	Register(app *App) error
}

func (a *App) RegisterModule(mod Module) error {
	if mod == nil {
		return ErrNilModule
	}
	return mod.Register(a)
}

func (a *App) Module(mod Module) *App {
	if a.initErr != nil {
		return a
	}
	if err := a.RegisterModule(mod); err != nil {
		a.initErr = err
	}
	return a
}
