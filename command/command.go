package command

import (
	"time"

	"github.com/sfossl/gda/types"
)

type CommandType int

const (
	CommandTypeChatInput CommandType = 1
	CommandTypeUser      CommandType = 2
	CommandTypeMessage   CommandType = 3
)

type OptionType int

const (
	OptionTypeSubcommand      OptionType = 1
	OptionTypeSubcommandGroup OptionType = 2
	OptionTypeString          OptionType = 3
	OptionTypeInteger         OptionType = 4
	OptionTypeBoolean         OptionType = 5
	OptionTypeUser            OptionType = 6
	OptionTypeChannel         OptionType = 7
	OptionTypeRole            OptionType = 8
	OptionTypeMentionable     OptionType = 9
	OptionTypeNumber          OptionType = 10
	OptionTypeAttachment      OptionType = 11
)

type ChannelType int

const (
	ChannelTypeGuildText          ChannelType = 0
	ChannelTypeDM                 ChannelType = 1
	ChannelTypeGuildVoice         ChannelType = 2
	ChannelTypeGroupDM            ChannelType = 3
	ChannelTypeGuildCategory      ChannelType = 4
	ChannelTypeGuildAnnouncement  ChannelType = 5
	ChannelTypeAnnouncementThread ChannelType = 10
	ChannelTypePublicThread       ChannelType = 11
	ChannelTypePrivateThread      ChannelType = 12
	ChannelTypeGuildStageVoice    ChannelType = 13
	ChannelTypeGuildDirectory     ChannelType = 14
	ChannelTypeGuildForum         ChannelType = 15
	ChannelTypeGuildMedia         ChannelType = 16
)

type Choice = types.Choice

type Option struct {
	Type         OptionType
	Name         string
	Description  string
	Required     bool
	Choices      []Choice
	Autocomplete bool
	ChannelTypes []int
	MinValue     *float64
	MaxValue     *float64
	MinLength    *int
	MaxLength    *int
}

type Subcommand struct {
	Name        string
	Description string
	Options     []Option
	Execute     any
}

type SubcommandGroup struct {
	Name        string
	Description string
	Subcommands []Subcommand
}

type CooldownScope int

const (
	CooldownScopeUser CooldownScope = iota
	CooldownScopeChannel
	CooldownScopeGuild
	CooldownScopeGlobal
)

type Command struct {
	Type                     CommandType
	Name                     string
	Description              string
	GuildID                  types.Snowflake
	DefaultMemberPermissions *types.Permission
	DMPermission             *bool
	NSFW                     bool
	Disabled                 bool
	Options                  []Option
	Subcommands              []Subcommand
	SubcommandGroups         []SubcommandGroup
	Execute                  any
	AutocompleteHandler      any
	Permissions              types.Permission
	GuildOnly                bool
	Cooldown                 time.Duration
	CooldownScope            CooldownScope
}
