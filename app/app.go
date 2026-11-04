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

type commandKey struct {
	name string
	typ  int
}

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
	commands       map[commandKey]*commandEntry
	components     map[string]middleware.Handler
	modals         map[string]middleware.Handler
	syncedHashes   map[string]string
	mu             sync.RWMutex
	syncMu         sync.Mutex
	running        atomic.Bool
	initErr        error
	server         *http.Server
	rootCtx        stdctx.Context
	rootCancel     stdctx.CancelFunc
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

	rootCtx, rootCancel := stdctx.WithCancel(stdctx.Background())
	app := &App{
		config:         cfg,
		REST:           restClient,
		Gateway:        gatewayClient,
		Dispatcher:     dispatcher,
		globalPipeline: middleware.NewPipeline(),
		commands:       make(map[commandKey]*commandEntry),
		components:     make(map[string]middleware.Handler),
		modals:         make(map[string]middleware.Handler),
		syncedHashes:   make(map[string]string),
		rootCtx:        rootCtx,
		rootCancel:     rootCancel,
	}

	app.initDispatcher()
	return app, nil
}

func (a *App) initDispatcher() {
	a.Dispatcher.OnInteractionCreate(func(ctx stdctx.Context, in *types.Interaction) {
		bgCtx, cancel := stdctx.WithTimeout(a.rootCtx, a.config.Timeout)
		gdkCtx := context.NewGatewayContext(bgCtx, in, a.REST, cancel)
		gdkCtx.Gateway = a.Gateway
		defer func() {
			if rec := recover(); rec != nil {
				cancel()
			}
		}()
		err := a.Handle(gdkCtx)
		if err != nil {
			cancel()
		}
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
	if typ == 0 {
		typ = int(command.CommandTypeChatInput)
	}
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.commands[commandKey{name: name, typ: typ}]
}

func (a *App) RegisterCommand(cmd command.Command, handler middleware.Handler, middlewares ...middleware.Middleware) error {
	if err := command.Validate(cmd); err != nil {
		return err
	}

	c := cmd
	if c.Type == 0 {
		c.Type = command.CommandTypeChatInput
	}
	key := commandKey{name: c.Name, typ: int(c.Type)}

	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.commands[key]; exists {
		return ErrDuplicateCommand
	}

	if cmd.Cooldown > 0 {
		var scope middleware.Scope
		switch cmd.CooldownScope {
		case command.CooldownScopeChannel:
			scope = middleware.ScopeChannel
		case command.CooldownScopeGuild:
			scope = middleware.ScopeGuild
		case command.CooldownScopeGlobal:
			scope = middleware.ScopeGlobal
		default:
			scope = middleware.ScopeUser
		}
		middlewares = append([]middleware.Middleware{
			middleware.Cooldown([]middleware.Policy{
				{
					Scope:    scope,
					Duration: cmd.Cooldown,
					Burst:    1,
				},
			}, middleware.WithKeyPrefix("cmd:"+cmd.Name)),
		}, middlewares...)
	}

	if cmd.Permissions != 0 {
		middlewares = append([]middleware.Middleware{
			middleware.RequireUser(cmd.Permissions),
		}, middlewares...)
	}

	p := middleware.NewPipeline(middlewares...)
	a.commands[key] = &commandEntry{cmd: c, handler: handler, pipeline: p}

	return nil
}

func (a *App) Command(cmd command.Command, args ...any) *App {
	if a.initErr != nil || cmd.Disabled {
		return a
	}
	var handler middleware.Handler
	var mws []middleware.Middleware

	if cmd.Execute != nil {
		if h, ok := cmd.Execute.(middleware.Handler); ok {
			handler = h
		} else if h, ok := cmd.Execute.(func(*context.Context) error); ok {
			handler = h
		}
	}

	for _, arg := range args {
		switch v := arg.(type) {
		case middleware.Handler:
			handler = v
		case func(*context.Context) error:
			handler = v
		case middleware.Middleware:
			mws = append(mws, v)
		case []middleware.Middleware:
			mws = append(mws, v...)
		}
	}

	if err := a.RegisterCommand(cmd, handler, mws...); err != nil {
		a.initErr = err
	}
	return a
}

func (a *App) Commands(cmds ...command.Command) *App {
	for _, cmd := range cmds {
		a.Command(cmd)
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

	if a.config.AutoDefer && isDeferrableInteraction(ctx.Interaction.Type) {
		timer := time.AfterFunc(a.config.AutoDeferDelay, func() {
			if !ctx.IsAcknowledged() {
				_ = ctx.Defer(false)
			}
		})
		defer timer.Stop()
	}

	switch ctx.Interaction.Type {
	case types.InteractionTypePing:
		return a.handlePing(ctx)
	case types.InteractionTypeApplicationCommand:
		return a.handleCommand(ctx)
	case types.InteractionTypeMessageComponent:
		return a.handleComponent(ctx)
	case types.InteractionTypeApplicationCommandAutocomplete:
		return a.handleAutocomplete(ctx)
	case types.InteractionTypeModalSubmit:
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

	targetHandler := entry.handler
	if subName := ctx.Subcommand(); subName != "" {
		if subHandler := findSubcommandHandler(entry.cmd, ctx.SubcommandGroup(), subName); subHandler != nil {
			targetHandler = subHandler
		}
	}

	cmdHandler := entry.pipeline.Wrap(targetHandler)
	return a.globalPipeline.Execute(ctx, cmdHandler)
}

func findSubcommandHandler(cmd command.Command, group, name string) middleware.Handler {
	if group != "" {
		for _, g := range cmd.SubcommandGroups {
			if g.Name == group {
				for _, s := range g.Subcommands {
					if s.Name == name && s.Execute != nil {
						if h, ok := toMiddlewareHandler(s.Execute); ok {
							return h
						}
					}
				}
			}
		}
	}
	for _, s := range cmd.Subcommands {
		if s.Name == name && s.Execute != nil {
			if h, ok := toMiddlewareHandler(s.Execute); ok {
				return h
			}
		}
	}
	return nil
}

func toMiddlewareHandler(h any) (middleware.Handler, bool) {
	if h == nil {
		return nil, false
	}
	switch fn := h.(type) {
	case middleware.Handler:
		return fn, true
	case func(*context.Context) error:
		return fn, true
	default:
		return nil, false
	}
}

func (a *App) handleAutocomplete(ctx *context.Context) error {
	if ctx.Interaction.Data == nil {
		return ErrCommandNotFound
	}

	entry := a.lookupCommand(ctx.Interaction.Data.Type, ctx.Interaction.Data.Name)
	if entry == nil {
		return ErrCommandNotFound
	}

	if handler, ok := toMiddlewareHandler(entry.cmd.AutocompleteHandler); ok {
		return handler(ctx)
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

func (a *App) String() string {
	if a == nil {
		return "<nil>"
	}
	return a.config.String()
}

func isDeferrableInteraction(t types.InteractionType) bool {
	switch t {
	case types.InteractionTypeApplicationCommand,
		types.InteractionTypeMessageComponent,
		types.InteractionTypeModalSubmit:
		return true
	default:
		return false
	}
}


