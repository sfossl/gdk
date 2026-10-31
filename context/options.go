package context

import (
	"strconv"
	"strings"

	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/defaults"
)

type Options struct {
	Type         any
	Scope        any
	Target       types.Snowflake
	Channel      types.Snowflake
	Role         types.Snowflake
	Message      types.Snowflake
	Content      string
	Reason       string
	Receipt      string
	Duration     any
	Count        int
	Ephemeral    bool
	Mention      *bool
	Typing       bool
	Embeds       []types.Embed
	Components   []types.Component
	Files        []types.File
	TTS          bool
	Deferred     bool
	Name         string
	Topic        string
	Slowmode     int
	Color        int
	Permissions  *types.Permission
	Hoist        *bool
	Mentionable  *bool
	AddRoles     []types.Snowflake
	RemoveRoles  []types.Snowflake
	Nickname     *string
	VoiceMute    *bool
	VoiceDeaf    *bool
	VoiceChannel *types.Snowflake
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
	case ModType:
		switch v {
		case ModKick:
			return "kick"
		case ModBan:
			return "ban"
		case ModUnban:
			return "unban"
		case ModTimeout:
			return "timeout"
		case ModRemoveTimeout:
			return "removetimeout"
		}
	case ChannelAction:
		switch v {
		case ChannelCreate:
			return "create"
		case ChannelEdit:
			return "edit"
		case ChannelDelete:
			return "delete"
		case ChannelPurge:
			return "purge"
		case ChannelLock:
			return "lock"
		}
	case RoleAction:
		switch v {
		case RoleCreate:
			return "create"
		case RoleEdit:
			return "edit"
		case RoleDelete:
			return "delete"
		}
	case MessageAction:
		switch v {
		case MsgPin:
			return "pin"
		case MsgUnpin:
			return "unpin"
		case MsgDelete:
			return "delete"
		case MsgCrosspost:
			return "crosspost"
		}
	case int:
		return strconv.Itoa(v)
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
	if o.Message != 0 {
		return o.Message
	}
	if c != nil && c.Interaction != nil && c.Interaction.Data != nil && c.Interaction.Data.TargetID != nil {
		return *c.Interaction.Data.TargetID
	}
	return 0
}
