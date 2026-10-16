package app

import (
	stdctx "context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sfossl/gda/gateway"
	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/command"
	"github.com/sfossl/gdk/context"
	"github.com/sfossl/gdk/middleware"
)

var (
	ErrCommandNotFound     = errors.New("command not found")
	ErrComponentNotFound   = errors.New("component handler not found")
	ErrModalNotFound       = errors.New("modal handler not found")
	ErrAlreadyRunning      = errors.New("application is already running")
	ErrInvalidSignature    = errors.New("invalid interaction signature")
	ErrPublicKeyRequired   = errors.New("public key required for webhook mode")
	ErrDuplicateCommand    = errors.New("command already registered")
)

type commandEntry struct {
	cmd      command.Command
	handler  middleware.Handler
	pipeline *middleware.Pipeline
}

type App struct {
	config         Config
	REST           *rest.Client
	Gateway        *gateway.Client
	Dispatcher     *gateway.EventDispatcher
	globalPipeline *middleware.Pipeline
	commands       map[string]*commandEntry
	components     map[string]middleware.Handler
	modals         map[string]middleware.Handler
	syncedHashes   map[string]string
	mu             sync.RWMutex
	running        atomic.Bool
	initErr        error
	server         *http.Server
}

func New(token string, appID types.Snowflake, opts ...Option) (*App, error) {
	cfg := DefaultConfig(token, appID)
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	restOpts := []rest.ClientOption{
		rest.WithAppID(cfg.AppID.String()),
	}
	restClient := rest.NewClient(cfg.Token, restOpts...)

	dispatcher := gateway.NewEventDispatcher()
	gatewayClient := gateway.NewClient(gateway.Config{
		Token:   cfg.Token,
		Intents: cfg.Intents,
	}, dispatcher)

	app := &App{
		config:         cfg,
		REST:           restClient,
		Gateway:        gatewayClient,
		Dispatcher:     dispatcher,
		globalPipeline: middleware.NewPipeline(),
		commands:       make(map[string]*commandEntry),
		components:     make(map[string]middleware.Handler),
		modals:         make(map[string]middleware.Handler),
		syncedHashes:   make(map[string]string),
	}

	app.initDispatcher()
	return app, nil
}

func (a *App) initDispatcher() {
	a.Dispatcher.OnInteractionCreate(func(ctx stdctx.Context, in *types.Interaction) {
		gdkCtx := context.NewGatewayContext(ctx, in, a.REST)
		gdkCtx.Gateway = a.Gateway
		_ = a.Handle(gdkCtx)
	})
}

func (a *App) Config() Config {
	return a.config
}

func (a *App) Use(middlewares ...middleware.Middleware) *App {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.globalPipeline.Use(middlewares...)
	return a
}

func (a *App) lookupCommand(typ int, name string) *commandEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.commands[name]
}

func (a *App) RegisterCommand(cmd command.Command, handler middleware.Handler, middlewares ...middleware.Middleware) error {
	if err := command.Validate(cmd); err != nil {
		return err
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.commands[cmd.Name]; exists {
		return ErrDuplicateCommand
	}

	if cmd.Cooldown > 0 {
		middlewares = append([]middleware.Middleware{
			middleware.Cooldown([]middleware.Policy{
				middleware.UserCooldown(cmd.Cooldown),
			}),
		}, middlewares...)
	}

	p := middleware.NewPipeline(middlewares...)
	c := cmd
	c.Type = command.CommandTypeChatInput
	a.commands[cmd.Name] = &commandEntry{cmd: c, handler: handler, pipeline: p}

	return nil
}

func (a *App) Command(cmd command.Command, handler middleware.Handler, middlewares ...middleware.Middleware) *App {
	if a.initErr != nil {
		return a
	}
	if err := a.RegisterCommand(cmd, handler, middlewares...); err != nil {
		a.initErr = err
	}
	return a
}

func (a *App) Component(customID string, handler middleware.Handler) *App {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.components[customID] = handler
	return a
}

func (a *App) Modal(customID string, handler middleware.Handler) *App {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.modals[customID] = handler
	return a
}

func (a *App) Handle(ctx *context.Context) error {
	if ctx == nil || ctx.Interaction == nil {
		return errors.New("interaction context cannot be nil")
	}

	if a.config.AutoDefer {
		timer := time.AfterFunc(a.config.AutoDeferDelay, func() {
			if !ctx.IsAcknowledged() {
				_ = ctx.Defer(false)
			}
		})
		defer timer.Stop()
	}

	switch ctx.Interaction.Type {
	case 1:
		return a.handlePing(ctx)
	case 2:
		return a.handleCommand(ctx)
	case 3:
		return a.handleComponent(ctx)
	case 4:
		return a.handleAutocomplete(ctx)
	case 5:
		return a.handleModal(ctx)
	default:
		return nil
	}
}

func (a *App) handlePing(ctx *context.Context) error {
	return ctx.Pong()
}

func (a *App) handleCommand(ctx *context.Context) error {
	if ctx.Interaction.Data == nil {
		return ErrCommandNotFound
	}

	entry := a.lookupCommand(ctx.Interaction.Data.Type, ctx.Interaction.Data.Name)
	if entry == nil {
		return ErrCommandNotFound
	}

	cmdHandler := entry.pipeline.Wrap(entry.handler)
	return a.globalPipeline.Execute(ctx, cmdHandler)
}

func (a *App) handleAutocomplete(ctx *context.Context) error {
	if ctx.Interaction.Data == nil {
		return ErrCommandNotFound
	}

	entry := a.lookupCommand(ctx.Interaction.Data.Type, ctx.Interaction.Data.Name)
	if entry == nil {
		return ErrCommandNotFound
	}

	if entry.cmd.AutocompleteHandler != nil {
		if fn, ok := entry.cmd.AutocompleteHandler.(func(*context.Context) error); ok {
			return fn(ctx)
		}
		if fn, ok := entry.cmd.AutocompleteHandler.(middleware.Handler); ok {
			return fn(ctx)
		}
	}

	return nil
}

func (a *App) handleComponent(ctx *context.Context) error {
	if ctx.Interaction.Data == nil {
		return ErrComponentNotFound
	}

	customID := ctx.Interaction.Data.CustomID
	a.mu.RLock()
	handler, ok := a.components[customID]
	a.mu.RUnlock()

	if !ok || handler == nil {
		return ErrComponentNotFound
	}

	return a.globalPipeline.Execute(ctx, handler)
}

func (a *App) handleModal(ctx *context.Context) error {
	if ctx.Interaction.Data == nil {
		return ErrModalNotFound
	}

	customID := ctx.Interaction.Data.CustomID
	a.mu.RLock()
	handler, ok := a.modals[customID]
	a.mu.RUnlock()

	if !ok || handler == nil {
		return ErrModalNotFound
	}

	return a.globalPipeline.Execute(ctx, handler)
}

func (a *App) Err() error {
	return a.initErr
}

func (a *App) Run(ctx stdctx.Context) error {
	if a.initErr != nil {
		return a.initErr
	}
	if a.running.Swap(true) {
		return ErrAlreadyRunning
	}
	defer a.running.Store(false)

	if a.config.AutoSync {
		if err := a.Sync(ctx); err != nil {
			return err
		}
	}

	switch a.config.Mode {
	case ModeWebhook:
		return a.runWebhook(ctx)
	default:
		return a.runGateway(ctx)
	}
}

