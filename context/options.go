package context

import (
	"strings"
	"time"

	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/defaults"
)

type Options struct {
	Type             any
	Scope            any
	Target           types.Snowflake
	Channel          types.Snowflake
	Role             types.Snowflake
	Message          types.Snowflake
	Content          string
	Reason           string
	Emoji            any
	Duration         any
	Count            int
	Ephemeral        bool
	Mention          *bool
	Typing           bool
	Embed            any
	Embeds           []types.Embed
	Components       []types.Component
	Files            []types.File
	TTS              bool
	Deferred         bool
	ComponentTimeout any
	DeleteAfter      any
	Name             string
	Topic            string
	Slowmode         any
	Color            int
	Permissions      *types.Permission
	Hoist            *bool
	Mentionable      *bool
	AddRoles         []types.Snowflake
	RemoveRoles      []types.Snowflake
	Nickname         *string
	VoiceMute        *bool
	VoiceDeaf        *bool
	VoiceChannel     *types.Snowflake
}

func (o Options) ResolveComponentTimeout() time.Duration {
	return parseDuration(o.ComponentTimeout)
}

func (o Options) ResolveDeleteAfter() time.Duration {
	return parseDuration(o.DeleteAfter)
}

func parseDuration(v any) time.Duration {
	switch d := v.(type) {
	case time.Duration:
		return d
	case int:
		return time.Duration(d) * time.Millisecond
	case int64:
		return time.Duration(d) * time.Millisecond
	case float64:
		return time.Duration(d * float64(time.Millisecond))
	case string:
		if parsed, err := time.ParseDuration(d); err == nil {
			return parsed
		}
	}
	return 0
}

const (
	DefaultDMReceipt = defaults.DMReceipt
)

func (o Options) ResolveScope() string {
	if o.Scope == nil {
		return "guild"
	}
	switch v := o.Scope.(type) {
	case string:
		return strings.ToLower(strings.TrimSpace(v))
	default:
		return "guild"
	}
}

func (o Options) ResolveType() string {
	if o.Type == nil {
		return ""
	}
	switch v := o.Type.(type) {
	case string:
		return strings.ToLower(strings.TrimSpace(v))
	}
	return ""
}

func (o Options) ResolveTarget(c *Context) types.Snowflake {
	if o.Target != 0 {
		return o.Target
	}
	if c != nil {
		return c.TargetID()
	}
	return 0
}

func (o Options) ResolveChannel(c *Context) types.Snowflake {
	if o.Channel != 0 {
		return o.Channel
	}
	if c != nil {
		return c.ChannelID
	}
	return 0
}

func (o Options) ResolveMessage(c *Context) types.Snowflake {
	if o.Target != 0 {
		return o.Target
	}
	if o.Message != 0 {
		return o.Message
	}
	if c != nil && c.Interaction != nil && c.Interaction.Data != nil && c.Interaction.Data.TargetID != nil {
		return *c.Interaction.Data.TargetID
	}
	return 0
}
