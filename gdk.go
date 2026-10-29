package gdk

import (
	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/app"
	"github.com/sfossl/gdk/command"
	"github.com/sfossl/gdk/component"
	"github.com/sfossl/gdk/context"
	"github.com/sfossl/gdk/middleware"
	"github.com/sfossl/gdk/util/time"
)

type Snowflake = types.Snowflake
type Permission = types.Permission
type Intent = types.Intent
type Embed = types.Embed
type Component = types.Component
type ComponentType = types.ComponentType
type Message = types.Message
type User = types.User
type Member = types.Member
type Role = types.Role
type Channel = types.Channel
type Guild = types.Guild
type ReadyEvent = types.ReadyEvent
type Interaction = types.Interaction
type InteractionData = types.InteractionData
type InteractionResponse = types.InteractionResponse
type InteractionResponseData = types.InteractionResponseData
type InteractionType = types.InteractionType
type InteractionCallbackType = types.InteractionCallbackType

type App = app.App
type Config = app.Config
type Mode = app.Mode
type AppOption = app.Option
type Module = app.Module
type SyncResult = app.SyncResult
type Context = context.Context
type OptionVal = context.OptionVal
type Options = context.Options
type Response = context.Response
type Modal = context.Modal
type ModType = context.ModType
type ModOptions = context.ModOptions
type ModAction = context.ModAction
type ChannelAction = context.ChannelAction
type ChannelOptions = context.ChannelOptions
type ChannelSuite = context.ChannelSuite
type RoleAction = context.RoleAction
type RoleOptions = context.RoleOptions
type RoleSuite = context.RoleSuite
type MessageAction = context.MessageAction
type MessageOptions = context.MessageOptions
type MessageSuite = context.MessageSuite
type UpdateOptions = context.UpdateOptions
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
type Handler = middleware.Handler
type Middleware = middleware.Middleware
type PanicHandler = middleware.PanicHandler
type GuardConfig = middleware.GuardConfig
type GuardOption = middleware.GuardOption
type CooldownScope = command.CooldownScope
type Scope = middleware.Scope
type Policy = middleware.Policy
type CooldownStore = middleware.CooldownStore
type CooldownConfig = middleware.CooldownConfig
type CooldownOption = middleware.CooldownOption
type CooldownError = middleware.CooldownError
type PermissionTarget = middleware.PermissionTarget
type PermissionConfig = middleware.PermissionConfig
type PermissionOption = middleware.PermissionOption
type PermissionError = middleware.PermissionError
type Pipeline = middleware.Pipeline

const (
	ModeGateway = app.ModeGateway
	ModeWebhook = app.ModeWebhook

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

	ModKick          = context.ModKick
	ModBan           = context.ModBan
	ModUnban         = context.ModUnban
	ModTimeout       = context.ModTimeout
	ModRemoveTimeout = context.ModRemoveTimeout

	ChannelCreate = context.ChannelCreate
	ChannelEdit   = context.ChannelEdit
	ChannelDelete = context.ChannelDelete
	ChannelPurge  = context.ChannelPurge
	ChannelLock   = context.ChannelLock

	RoleCreate = context.RoleCreate
	RoleEdit   = context.RoleEdit
	RoleDelete = context.RoleDelete

	MsgPin       = context.MsgPin
	MsgUnpin     = context.MsgUnpin
	MsgDelete    = context.MsgDelete
	MsgCrosspost = context.MsgCrosspost

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

	CooldownScopeUser    = command.CooldownScopeUser
	CooldownScopeChannel = command.CooldownScopeChannel
	CooldownScopeGuild   = command.CooldownScopeGuild
	CooldownScopeGlobal  = command.CooldownScopeGlobal

	ScopeUser    = middleware.ScopeUser
	ScopeChannel = middleware.ScopeChannel
	ScopeGuild   = middleware.ScopeGuild
	ScopeGlobal  = middleware.ScopeGlobal

	TargetUser = middleware.TargetUser
	TargetBot  = middleware.TargetBot

	PermissionAdministrator         = types.PermissionAdministrator
	PermissionManageGuild           = types.PermissionManageGuild
	PermissionManageRoles           = types.PermissionManageRoles
	PermissionManageChannels        = types.PermissionManageChannels
	PermissionKickMembers           = types.PermissionKickMembers
	PermissionBanMembers            = types.PermissionBanMembers
	PermissionModerateMembers       = types.PermissionModerateMembers
	PermissionManageMessages        = types.PermissionManageMessages
	PermissionSendMessages          = types.PermissionSendMessages
	PermissionViewChannel           = types.PermissionViewChannel

	IntentGuilds                    = types.IntentGuilds
	IntentGuildMembers              = types.IntentGuildMembers
	IntentGuildMessages             = types.IntentGuildMessages
	IntentMessageContent            = types.IntentMessageContent
	IntentsDefault                  = types.IntentsDefault
	IntentsNonPrivileged            = types.IntentsNonPrivileged

	InteractionTypePing                           = types.InteractionTypePing
	InteractionTypeApplicationCommand             = types.InteractionTypeApplicationCommand
	InteractionTypeMessageComponent               = types.InteractionTypeMessageComponent
	InteractionTypeApplicationCommandAutocomplete = types.InteractionTypeApplicationCommandAutocomplete
	InteractionTypeModalSubmit                    = types.InteractionTypeModalSubmit

	CallbackTypePong                                    = types.CallbackTypePong
	CallbackTypeChannelMessageWithSource                = types.CallbackTypeChannelMessageWithSource
	CallbackTypeDeferredChannelMessageWithSource        = types.CallbackTypeDeferredChannelMessageWithSource
	CallbackTypeDeferredUpdateMessage                   = types.CallbackTypeDeferredUpdateMessage
	CallbackTypeUpdateMessage                           = types.CallbackTypeUpdateMessage
	CallbackTypeApplicationCommandAutocompleteResult    = types.CallbackTypeApplicationCommandAutocompleteResult
	CallbackTypeModal                                   = types.CallbackTypeModal

	ComponentTypeActionRow         = types.ComponentTypeActionRow
	ComponentTypeButton            = types.ComponentTypeButton
	ComponentTypeStringSelect      = types.ComponentTypeStringSelect
	ComponentTypeTextInput         = types.ComponentTypeTextInput
	ComponentTypeUserSelect        = types.ComponentTypeUserSelect
	ComponentTypeRoleSelect        = types.ComponentTypeRoleSelect
	ComponentTypeMentionableSelect = types.ComponentTypeMentionableSelect
	ComponentTypeChannelSelect     = types.ComponentTypeChannelSelect
)

var (
	ParseSnowflake = types.ParseSnowflake

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
	ErrContextMenuHasDescription = command.ErrContextMenuHasDescription
	ErrContextMenuHasOptions     = command.ErrContextMenuHasOptions

	ErrNotInGuild               = context.ErrNotInGuild
	ErrNoRESTClient             = context.ErrNoRESTClient
	ErrCannotModerateSelf       = context.ErrCannotModerateSelf
	ErrCannotModerateOwner      = context.ErrCannotModerateOwner
	ErrTargetOutranksUser       = context.ErrTargetOutranksUser
	ErrTargetOutranksBot        = context.ErrTargetOutranksBot
	ErrInvalidTarget            = context.ErrInvalidTarget
	ErrInvalidModType           = context.ErrInvalidModType
	ErrAlreadyResponded         = context.ErrAlreadyResponded
	ErrModalAfterDefer          = context.ErrModalAfterDefer
	ErrNoInteraction            = context.ErrNoInteraction
	ErrTimeoutTooLong           = context.ErrTimeoutTooLong
	ErrTimeoutTooShort          = context.ErrTimeoutTooShort
	ErrBanDeleteTooLong         = context.ErrBanDeleteTooLong
	ErrInvalidChannelID         = context.ErrInvalidChannelID
	ErrChannelNameRequired      = context.ErrChannelNameRequired
	ErrSlowmodeTooLong          = context.ErrSlowmodeTooLong
	ErrSlowmodeNegative         = context.ErrSlowmodeNegative
	ErrInvalidPurgeCount        = context.ErrInvalidPurgeCount
	ErrInvalidChannelAction     = context.ErrInvalidChannelAction
	ErrInvalidRoleID            = context.ErrInvalidRoleID
	ErrRoleNotFound             = context.ErrRoleNotFound
	ErrRoleOutranksBot          = context.ErrRoleOutranksBot
	ErrRoleOutranksUser         = context.ErrRoleOutranksUser
	ErrCannotModifyEveryoneRole = context.ErrCannotModifyEveryoneRole
	ErrInvalidRoleAction        = context.ErrInvalidRoleAction
	ErrInvalidMessageID         = context.ErrInvalidMessageID
	ErrInvalidMessageAction     = context.ErrInvalidMessageAction
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
	ErrSelectMissingOptions    = component.ErrSelectMissingOptions
	ErrSelectTooManyOptions    = component.ErrSelectTooManyOptions

	Guard            = middleware.Guard
	WithPanicHandler = middleware.WithPanicHandler
	WithMessage      = middleware.WithMessage
	WithEphemeral    = middleware.WithEphemeral

	Cooldown                = middleware.Cooldown
	UserCooldown            = middleware.UserCooldown
	ChannelCooldown         = middleware.ChannelCooldown
	GuildCooldown           = middleware.GuildCooldown
	GlobalCooldown          = middleware.GlobalCooldown
	NewCooldownStore        = middleware.NewCooldownStore
	WithCooldownStore       = middleware.WithCooldownStore
	WithCooldownAutoRespond = middleware.WithCooldownAutoRespond
	WithCooldownMessage     = middleware.WithCooldownMessage
	WithKeyPrefix           = middleware.WithKeyPrefix

	RequireUser               = middleware.RequireUser
	RequireBot                = middleware.RequireBot
	RequirePermissions        = middleware.RequirePermissions
	PermissionNames           = middleware.PermissionNames
	WithPermissionAutoRespond = middleware.WithPermissionAutoRespond
	WithPermissionMessage     = middleware.WithPermissionMessage

	NewPipeline = middleware.NewPipeline
	Chain       = middleware.Chain

	ErrPanicRecovered      = middleware.ErrPanicRecovered
	ErrOnCooldown          = middleware.ErrOnCooldown
	ErrInvalidCooldown     = middleware.ErrInvalidCooldown
	ErrMissingGuildInfo    = middleware.ErrMissingGuildInfo
	ErrMissingPermissions  = middleware.ErrMissingPermissions
	ErrNoMemberPermissions = middleware.ErrNoMemberPermissions
	ErrNoBotPermissions    = middleware.ErrNoBotPermissions

	NewApp        = app.New
	DefaultConfig = app.DefaultConfig

	WithMode      = app.WithMode
	WithIntents   = app.WithIntents
	WithDevMode   = app.WithDevMode
	WithAutoDefer = app.WithAutoDefer
	WithAutoSync  = app.WithAutoSync
	WithTimeout   = app.WithTimeout
	WithWebhook   = app.WithWebhook

	ErrTokenRequired     = app.ErrTokenRequired
	ErrAppIDRequired     = app.ErrAppIDRequired
	ErrCommandNotFound   = app.ErrCommandNotFound
	ErrComponentNotFound = app.ErrComponentNotFound
	ErrModalNotFound     = app.ErrModalNotFound
	ErrAlreadyRunning    = app.ErrAlreadyRunning
	ErrInvalidSignature  = app.ErrInvalidSignature
	ErrPublicKeyRequired = app.ErrPublicKeyRequired
	ErrDuplicateCommand  = app.ErrDuplicateCommand
	ErrDevGuildRequired  = app.ErrDevGuildRequired
	ErrNilModule         = app.ErrNilModule
)
