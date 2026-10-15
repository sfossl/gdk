package gdk

import (
	"github.com/sfossl/gdk/command"
	"github.com/sfossl/gdk/context"
)

type Context = context.Context
type OptionVal = context.OptionVal
type Response = context.Response
type Command = command.Command
type CommandType = command.CommandType
type Option = command.Option
type OptionType = command.OptionType
type Choice = command.Choice
type Subcommand = command.Subcommand
type SubcommandGroup = command.SubcommandGroup

const (
	CommandTypeChatInput = command.CommandTypeChatInput
	CommandTypeUser      = command.CommandTypeUser
	CommandTypeMessage   = command.CommandTypeMessage

	OptionTypeSubcommand      = command.OptionTypeSubcommand
	OptionTypeSubcommandGroup = command.OptionTypeSubcommandGroup
	OptionTypeString          = command.OptionTypeString
	OptionTypeInteger         = command.OptionTypeInteger
	OptionTypeBoolean         = command.OptionTypeBoolean
	OptionTypeUser            = command.OptionTypeUser
	OptionTypeChannel         = command.OptionTypeChannel
	OptionTypeRole            = command.OptionTypeRole
	OptionTypeMentionable     = command.OptionTypeMentionable
	OptionTypeNumber          = command.OptionTypeNumber
	OptionTypeAttachment      = command.OptionTypeAttachment
)

var (
	StringOption      = command.StringOption
	IntOption         = command.IntOption
	NumberOption      = command.NumberOption
	BoolOption        = command.BoolOption
	UserOption        = command.UserOption
	ChannelOption     = command.ChannelOption
	RoleOption        = command.RoleOption
	MentionableOption = command.MentionableOption
	AttachmentOption  = command.AttachmentOption
	ValidateCommand   = command.Validate

	ErrNotInGuild          = context.ErrNotInGuild
	ErrNoRESTClient        = context.ErrNoRESTClient
	ErrCannotModerateSelf  = context.ErrCannotModerateSelf
	ErrCannotModerateOwner = context.ErrCannotModerateOwner
	ErrTargetOutranksUser  = context.ErrTargetOutranksUser
	ErrTargetOutranksBot   = context.ErrTargetOutranksBot
	ErrAlreadyResponded    = context.ErrAlreadyResponded
	ErrNoInteraction       = context.ErrNoInteraction
)

