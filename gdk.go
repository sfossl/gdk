package gdk

import (
	"github.com/sfossl/gdk/command"
	"github.com/sfossl/gdk/component"
	"github.com/sfossl/gdk/context"
	"github.com/sfossl/gdk/util/time"
)

type Context = context.Context
type OptionVal = context.OptionVal
type Response = context.Response
type ModAction = context.ModAction
type ChannelAction = context.ChannelAction
type RoleAction = context.RoleAction
type MessageAction = context.MessageAction
type TimeItem = timeutil.Item
type Command = command.Command
type CommandType = command.CommandType
type Option = command.Option
type OptionType = command.OptionType
type Choice = command.Choice
type Subcommand = command.Subcommand
type SubcommandGroup = command.SubcommandGroup
type EmbedBuilder = component.EmbedBuilder
type ButtonOpt = component.ButtonOpt
type SelectOpt = component.SelectOpt
type TextInputOpt = component.TextInputOpt

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
	MaxSlowmodeSeconds = context.MaxSlowmodeSeconds

	ColorBlurple = component.ColorBlurple
	ColorGreen   = component.ColorGreen
	ColorYellow  = component.ColorYellow
	ColorFuchsia = component.ColorFuchsia
	ColorRed     = component.ColorRed
	ColorWhite   = component.ColorWhite
	ColorBlack   = component.ColorBlack
	ColorDark    = component.ColorDark
	ColorNavy    = component.ColorNavy
	ColorAqua    = component.ColorAqua
	ColorTeal    = component.ColorTeal
	ColorOlive   = component.ColorOlive
	ColorLime    = component.ColorLime
	ColorOrange  = component.ColorOrange
	ColorMaroon  = component.ColorMaroon
	ColorPurple  = component.ColorPurple

	ButtonStylePrimary   = component.ButtonStylePrimary
	ButtonStyleSecondary = component.ButtonStyleSecondary
	ButtonStyleSuccess   = component.ButtonStyleSuccess
	ButtonStyleDanger    = component.ButtonStyleDanger
	ButtonStyleLink      = component.ButtonStyleLink

	TextInputStyleShort     = component.TextInputStyleShort
	TextInputStyleParagraph = component.TextInputStyleParagraph

	MaxActionRows       = component.MaxActionRows
	MaxButtonsPerRow    = component.MaxButtonsPerRow
	MaxSelectsPerRow    = component.MaxSelectsPerRow
	MaxTextInputsPerRow = component.MaxTextInputsPerRow
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

	ErrNotInGuild               = context.ErrNotInGuild
	ErrNoRESTClient             = context.ErrNoRESTClient
	ErrCannotModerateSelf       = context.ErrCannotModerateSelf
	ErrCannotModerateOwner      = context.ErrCannotModerateOwner
	ErrTargetOutranksUser       = context.ErrTargetOutranksUser
	ErrTargetOutranksBot        = context.ErrTargetOutranksBot
	ErrInvalidTarget            = context.ErrInvalidTarget
	ErrAlreadyResponded         = context.ErrAlreadyResponded
	ErrNoInteraction            = context.ErrNoInteraction
	ErrTimeoutTooLong           = context.ErrTimeoutTooLong
	ErrTimeoutTooShort          = context.ErrTimeoutTooShort
	ErrBanDeleteTooLong         = context.ErrBanDeleteTooLong
	ErrInvalidChannelID         = context.ErrInvalidChannelID
	ErrChannelNameRequired      = context.ErrChannelNameRequired
	ErrSlowmodeTooLong          = context.ErrSlowmodeTooLong
	ErrSlowmodeNegative         = context.ErrSlowmodeNegative
	ErrInvalidRoleID            = context.ErrInvalidRoleID
	ErrRoleNotFound             = context.ErrRoleNotFound
	ErrRoleOutranksBot          = context.ErrRoleOutranksBot
	ErrRoleOutranksUser         = context.ErrRoleOutranksUser
	ErrCannotModifyEveryoneRole = context.ErrCannotModifyEveryoneRole
	ErrInvalidMessageID         = context.ErrInvalidMessageID
	ErrEmojiRequired            = context.ErrEmojiRequired

	Time                = timeutil.New
	FormatTime          = timeutil.Format
	ParseDuration       = timeutil.ParseDuration
	FormatDuration      = timeutil.FormatDuration
	FormatDurationHuman = timeutil.FormatDurationHuman
	ErrInvalidDuration  = timeutil.ErrInvalidDuration
	ErrEmptyDuration    = timeutil.ErrEmptyDuration
	ErrUnknownUnit      = timeutil.ErrUnknownUnit
	ErrInvalidTime      = timeutil.ErrInvalidTime

	NewEmbed     = component.NewEmbed
	RGB          = component.RGB
	Hex          = component.Hex
	ResolveColor = component.ResolveColor

	Primary            = component.Primary
	Secondary          = component.Secondary
	Success            = component.Success
	Danger             = component.Danger
	Link               = component.Link
	WithButtonEmoji    = component.WithButtonEmoji
	WithButtonDisabled = component.WithButtonDisabled

	StringSelect       = component.StringSelect
	UserSelect         = component.UserSelect
	RoleSelect         = component.RoleSelect
	MentionableSelect  = component.MentionableSelect
	ChannelSelect      = component.ChannelSelect
	WithRange          = component.WithRange
	WithSelectDisabled = component.WithSelectDisabled
	WithChannelTypes   = component.WithChannelTypes
	SelectOption       = component.SelectOption
	WithOptionEmoji    = component.WithOptionEmoji
	WithOptionDefault  = component.WithOptionDefault

	TextInputShort     = component.TextInputShort
	TextInputParagraph = component.TextInputParagraph
	WithPlaceholder    = component.WithPlaceholder
	WithValue          = component.WithValue
	WithMinLength      = component.WithMinLength
	WithMaxLength      = component.WithMaxLength
	WithRequired       = component.WithRequired

	ActionRow    = component.ActionRow
	Row          = component.Row
	PackRows     = component.PackRows
	ValidateRows = component.ValidateRows

	ErrInvalidColor            = component.ErrInvalidColor
	ErrColorOutOfRange         = component.ErrColorOutOfRange
	ErrEmbedTitleTooLong       = component.ErrEmbedTitleTooLong
	ErrEmbedDescriptionTooLong = component.ErrEmbedDescriptionTooLong
	ErrEmbedTooManyFields      = component.ErrEmbedTooManyFields
	ErrEmbedFieldNameTooLong   = component.ErrEmbedFieldNameTooLong
	ErrEmbedFieldValueTooLong  = component.ErrEmbedFieldValueTooLong
	ErrEmbedFieldMissingName   = component.ErrEmbedFieldMissingName
	ErrEmbedFieldMissingValue  = component.ErrEmbedFieldMissingValue
	ErrEmbedFooterTooLong      = component.ErrEmbedFooterTooLong
	ErrEmbedAuthorTooLong      = component.ErrEmbedAuthorTooLong
	ErrEmbedTotalLengthTooLong = component.ErrEmbedTotalLengthTooLong
	ErrTooManyActionRows       = component.ErrTooManyActionRows
	ErrRowTooManyButtons       = component.ErrRowTooManyButtons
	ErrRowMixedComponents      = component.ErrRowMixedComponents
	ErrRowMultipleSelects      = component.ErrRowMultipleSelects
	ErrRowMultipleInputs       = component.ErrRowMultipleInputs
	ErrEmptyActionRow          = component.ErrEmptyActionRow
	ErrNonActionRowAtRoot      = component.ErrNonActionRowAtRoot
)
