package context

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sfossl/gda/gateway"
	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

type Context struct {
	ctx             context.Context
	Interaction     *types.Interaction
	REST            *rest.Client
	Gateway         *gateway.Client
	ID              types.Snowflake
	AppID           types.Snowflake
	GuildID         *types.Snowflake
	ChannelID       types.Snowflake
	Author          *types.User
	InvokerMember   *types.Member
	responded       atomic.Bool
	deferred        atomic.Bool
	gatewayMode     bool
	responseChan    chan types.InteractionResponse
	storeMu         sync.RWMutex
	store           map[string]any
	optsMu          sync.RWMutex
	optsMap         map[string]OptionVal
	subcommand      string
	subcommandGroup string
	cacheMu         sync.Mutex
	cachedGuild     *types.Guild
	cachedRoles     []types.Role
	cachedBotMem    *types.Member
	lifecycleMu      sync.Mutex
	isDeleted        bool
	compTimeoutTimer *time.Timer
	deleteAfterTimer *time.Timer
	activeComponents []types.Component
	msgID            types.Snowflake
	lastPurged       int
	cancelFunc       context.CancelFunc
}

func New(ctx context.Context, in *types.Interaction, restClient *rest.Client, cancel ...context.CancelFunc) *Context {
	var cFunc context.CancelFunc
	if len(cancel) > 0 {
		cFunc = cancel[0]
	}
	c := &Context{
		ctx:          ctx,
		Interaction:  in,
		REST:         restClient,
		responseChan: make(chan types.InteractionResponse, 1),
		store:        make(map[string]any),
		cancelFunc:   cFunc,
	}
	c.initFields(in)
	return c
}

func NewGatewayContext(ctx context.Context, in *types.Interaction, restClient *rest.Client, cancel ...context.CancelFunc) *Context {
	var cFunc context.CancelFunc
	if len(cancel) > 0 {
		cFunc = cancel[0]
	}
	c := &Context{
		ctx:         ctx,
		Interaction: in,
		REST:        restClient,
		gatewayMode: true,
		store:       make(map[string]any),
		cancelFunc:  cFunc,
	}
	c.initFields(in)
	return c
}

func (c *Context) initFields(in *types.Interaction) {
	if in == nil {
		return
	}
	c.ID = in.ID
	c.AppID = in.AppID
	c.GuildID = in.GuildID

	if in.ChannelID != nil {
		c.ChannelID = *in.ChannelID
	}

	if in.User != nil {
		c.Author = in.User
	} else if in.Member != nil && in.Member.User != nil {
		c.Author = in.Member.User
	}

	if in.Member != nil && in.GuildID != nil && in.Member.GuildID == 0 {
		in.Member.GuildID = *in.GuildID
	}
	c.InvokerMember = in.Member

	if in.Message != nil && in.Message.ID != 0 {
		c.msgID = in.Message.ID
	}

	if in.Data != nil && len(in.Data.Options) > 0 {
		first := in.Data.Options[0]
		switch first.Type {
		case int(types.OptionTypeSubcommand):
			c.subcommand = first.Name
		case int(types.OptionTypeSubcommandGroup):
			c.subcommandGroup = first.Name
			if len(first.Options) > 0 && first.Options[0].Type == int(types.OptionTypeSubcommand) {
				c.subcommand = first.Options[0].Name
			}
		}
	}
}

func (c *Context) Subcommand() string {
	return c.subcommand
}

func (c *Context) SubcommandGroup() string {
	return c.subcommandGroup
}

func (c *Context) LastPurged() int {
	return c.lastPurged
}

func (c *Context) Values() []string {
	if c.Interaction != nil && c.Interaction.Data != nil {
		return c.Interaction.Data.Values
	}
	return nil
}

func (c *Context) Value() string {
	vals := c.Values()
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func (c *Context) Input(name string) string {
	return c.Opt(name).String()
}

func (c *Context) Context() context.Context {
	if c.ctx == nil {
		return context.Background()
	}
	return c.ctx
}

func (c *Context) Cancel() {
	if c.cancelFunc != nil {
		c.cancelFunc()
	}
}

func (c *Context) ResponseChannel() <-chan types.InteractionResponse {
	return c.responseChan
}

func (c *Context) IsAcknowledged() bool {
	return c.responded.Load()
}

func (c *Context) IsDeferred() bool {
	return c.deferred.Load()
}

func (c *Context) Set(key string, val any) {
	c.storeMu.Lock()
	defer c.storeMu.Unlock()
	c.store[key] = val
}

func (c *Context) Get(key string) any {
	c.storeMu.RLock()
	defer c.storeMu.RUnlock()
	return c.store[key]
}

func (c *Context) Latency() time.Duration {
	if c.Gateway != nil {
		return c.Gateway.Latency()
	}
	return 0
}

func (c *Context) FocusedOption() *types.ApplicationCommandOption {
	if c == nil || c.Interaction == nil || c.Interaction.Data == nil {
		return nil
	}
	return findFocusedOption(c.Interaction.Data.Options)
}

func (c *Context) FocusedValue() string {
	f := c.FocusedOption()
	if f == nil || f.Value == nil {
		return ""
	}
	if s, ok := f.Value.(string); ok {
		return s
	}
	return fmt.Sprint(f.Value)
}

func (c *Context) FocusedName() string {
	f := c.FocusedOption()
	if f == nil {
		return ""
	}
	return f.Name
}

func findFocusedOption(opts []types.ApplicationCommandOption) *types.ApplicationCommandOption {
	for i := range opts {
		if opts[i].Focused {
			return &opts[i]
		}
		if len(opts[i].Options) > 0 {
			if f := findFocusedOption(opts[i].Options); f != nil {
				return f
			}
		}
	}
	return nil
}
