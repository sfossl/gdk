package context

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

type Context struct {
	ctx          context.Context
	Interaction  *types.Interaction
	REST         *rest.Client
	ID           types.Snowflake
	AppID        types.Snowflake
	GuildID      *types.Snowflake
	ChannelID    types.Snowflake
	Author       *types.User
	Member       *types.Member
	responded    atomic.Bool
	deferred     atomic.Bool
	gatewayMode  bool
	responseChan chan types.InteractionResponse
	storeMu      sync.RWMutex
	store        map[string]any
	optsMu       sync.RWMutex
	optsMap      map[string]OptionVal
}

func New(ctx context.Context, in *types.Interaction, restClient *rest.Client) *Context {
	c := &Context{
		ctx:          ctx,
		Interaction:  in,
		REST:         restClient,
		responseChan: make(chan types.InteractionResponse, 1),
		store:        make(map[string]any),
	}
	c.initFields(in)
	return c
}

func NewGatewayContext(ctx context.Context, in *types.Interaction, restClient *rest.Client) *Context {
	c := &Context{
		ctx:         ctx,
		Interaction: in,
		REST:        restClient,
		gatewayMode: true,
		store:       make(map[string]any),
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

	c.Member = in.Member
}

func (c *Context) Context() context.Context {
	if c.ctx == nil {
		return context.Background()
	}
	return c.ctx
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

