package gdk

import (
	"github.com/sfossl/gdk/command"
	"github.com/sfossl/gdk/context"
	"github.com/sfossl/gdk/util/time"
)

type Context = context.Context
type OptionVal = context.OptionVal
type Response = context.Response
type ModAction = context.ModAction
type TimeItem = timeutil.Item
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

	MaxTimeoutDuration = context.MaxTimeoutDuration
	MaxBanDeleteSec    = context.MaxBanDeleteSec
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
	ErrInvalidTarget       = context.ErrInvalidTarget
	ErrAlreadyResponded    = context.ErrAlreadyResponded
	ErrNoInteraction       = context.ErrNoInteraction
	ErrTimeoutTooLong      = context.ErrTimeoutTooLong
	ErrTimeoutTooShort     = context.ErrTimeoutTooShort
	ErrBanDeleteTooLong    = context.ErrBanDeleteTooLong

	Time                 = timeutil.New
	FormatTime           = timeutil.Format
	ParseDuration        = timeutil.ParseDuration
	FormatDuration       = timeutil.FormatDuration
	FormatDurationHuman  = timeutil.FormatDurationHuman
	ErrInvalidDuration   = timeutil.ErrInvalidDuration
	ErrEmptyDuration     = timeutil.ErrEmptyDuration
	ErrUnknownUnit       = timeutil.ErrUnknownUnit
	ErrInvalidTime       = timeutil.ErrInvalidTime
)

