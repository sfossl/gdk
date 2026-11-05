[Docs Overview](../README.md#documentation) / API Reference

# Reference

Complete reference of exported packages, types, constructors and functions in `github.com/sfossl/gdk`.

### Documentation Guides
[Getting Started](getting-started.md) | [Commands](commands.md) | [Context](context.md) | [Entity Modules](modules/index.md) | [Components](components.md) | [Middleware](middleware.md) | [Synchronization](sync.md) | [Transports](transports.md) | [Utilities](utilities.md)

## Application

### Constructors & Configuration
```go
func NewApp(token string, appID types.Snowflake, opts ...AppOption) (*App, error)
func DefaultConfig(token string, appID types.Snowflake) Config
```

### Functional Options
```go
func WithMode(mode Mode) AppOption
func WithIntents(intents types.Intent) AppOption
func WithDevMode(devMode bool, devGuildID ...types.Snowflake) AppOption
func WithAutoDefer(autoDefer bool, delay ...time.Duration) AppOption
func WithTimeout(timeout time.Duration) AppOption
func WithWebhook(addr, path, publicKey string) AppOption
func WithAutoSync(autoSync bool) AppOption
```

### `type App` Methods
```go
func (a *App) Run(ctx context.Context) error
func (a *App) Close() error
func (a *App) RegisterCommand(cmd Command, handler Handler, middlewares ...Middleware) error
func (a *App) Command(cmd Command, args ...any) *App
func (a *App) Commands(cmds ...Command) *App
func (a *App) MountRegistered() *App
func (a *App) Component(customID string, handler Handler) *App
func (a *App) Modal(customID string, handler Handler) *App
func (a *App) Handle(interactionType int, customID string, handler Handler) *App
func (a *App) RegisterModule(m Module) error
func (a *App) Module(m Module) *App
func (a *App) Use(middlewares ...Middleware) *App
func (a *App) Sync(ctx context.Context, force ...bool) error
func (a *App) SyncWithResult(ctx context.Context, force ...bool) (*SyncResult, error)
func (a *App) Config() Config
func (a *App) Err() error
func (a *App) String() string
```

### Global Registration
```go
func Register(cmd Command)
func RegisteredCommands() []Command
func RegisterComponent(customID string, handler Handler)
func RegisteredComponents() []ComponentRegistration
func RegisterModal(customID string, handler Handler)
func RegisteredModals() []ModalRegistration

type ComponentRegistration struct {
	CustomID string
	Handler  Handler
}

type ModalRegistration struct {
	CustomID string
	Handler  Handler
}
```

## Commands

### `type Command`
```go
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
```

### `type CommandType`
```go
type CommandType int

const (
	CommandTypeChatInput CommandType = 1 // Slash command
	CommandTypeUser      CommandType = 2 // User context menu
	CommandTypeMessage   CommandType = 3 // Message context menu
)
```

### `type CooldownScope`
```go
type CooldownScope int

const (
	CooldownScopeUser CooldownScope = iota
	CooldownScopeChannel
	CooldownScopeGuild
	CooldownScopeGlobal
)
```

### `type Choice`
```go
type Choice = types.Choice
```

### Option Constructors
```go
func StringOption(name, description string) Option
func IntOption(name, description string) Option
func NumberOption(name, description string) Option
func BoolOption(name, description string) Option
func UserOption(name, description string) Option
func ChannelOption(name, description string, channelTypes ...int) Option
func RoleOption(name, description string) Option
func MentionableOption(name, description string) Option
func AttachmentOption(name, description string) Option
```

### Option Chained Methods
```go
func (o Option) WithRequired() Option
func (o Option) WithChoices(choices ...Choice) Option
func (o Option) WithChoice(name string, value any) Option
func (o Option) WithAutocomplete() Option
func (o Option) WithMinMax(min, max float64) Option
func (o Option) WithLength(min, max int) Option
```

### Validation
```go
func ValidateCommand(cmd Command) error
```

## Context

### Universal `Options` Struct
```go
type Options struct {
	Type         any
	Scope        any
	Target       types.Snowflake
	Channel      types.Snowflake
	Role         types.Snowflake
	Message      types.Snowflake
	Content      string
	Reason       string
	Emoji        any
	Duration     any
	Count        int
	Ephemeral    bool
	Mention      *bool
	Typing       bool
	Embed        any
	Embeds       []types.Embed
	Components   []types.Component
	Files            []types.File
	TTS              bool
	Deferred         bool
	ComponentTimeout any
	DeleteAfter      any
	Name             string
	Topic        string
	Slowmode     any
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

type MemberOptions struct {
	Type         ModType
	Target       types.Snowflake
	Reason       string
	Duration     any
	Nickname     *string
	AddRoles     []types.Snowflake
	RemoveRoles  []types.Snowflake
	Role         types.Snowflake
	VoiceMute    *bool
	VoiceDeaf    *bool
	VoiceChannel *types.Snowflake
}

type ChannelOptions struct {
	Type       ChannelAction
	ChannelID  types.Snowflake
	Name       string
	Topic      string
	Slowmode   any
	PurgeCount int
	Locked     bool
	ParentID   *types.Snowflake
	Reason     string
}

type RoleOptions struct {
	Type        RoleAction
	RoleID      types.Snowflake
	Name        string
	Color       any
	Permissions *types.Permission
	Hoist       *bool
	Mentionable *bool
	Target      types.Snowflake
	Reason      string
}

type MessageOptions struct {
	Type       MessageAction
	MessageID  types.Snowflake
	ChannelID  types.Snowflake
	Content    string
	Embeds     []types.Embed
	Components []types.Component
	Emoji      []string
	UserID     types.Snowflake
	PurgeCount int
	Reason     string
}
```

### The 6 Unified Methods
```go
// respond
func (c *Context) Respond(args ...any) (*types.Message, error)
func (c *Context) Defer(ephemeral ...bool) error
func (c *Context) UpdateMessage(args ...any) (*types.Message, error)
func (c *Context) DisableComponents() error
func (c *Context) DeferUpdate() error
func (c *Context) Modal(m Modal) error
func (c *Context) Autocomplete(choices []Choice) error
func (c *Context) Pong() error

// member
func (c *Context) Member(args ...any) error
func (c *Context) Mod(args ...any) error // alias for Member
func (c *Context) Update(args ...any) error

type ModType int

const (
	ModKick ModType = iota + 1
	ModBan
	ModUnban
	ModTimeout
	ModRemoveTimeout
	ModAddRole
	ModRemoveRole
	ModNick
	ModResetNick
	ModMute
	ModDeaf
	ModMove
	ModDisconnect
)

// channel
func (c *Context) Channel(args ...any) error
func (c *Context) ManageChannel(args ...any) error // alias for Channel

type ChannelAction int

const (
	ChannelCreate ChannelAction = iota + 1
	ChannelEdit
	ChannelDelete
	ChannelPurge
	ChannelLock
)

// role
func (c *Context) Role(args ...any) (*types.Role, error)
func (c *Context) ManageRole(args ...any) (*types.Role, error) // alias for Role

type RoleAction int

const (
	RoleCreate RoleAction = iota + 1
	RoleEdit
	RoleDelete
)

// message
func (c *Context) Message(args ...any) error
func (c *Context) ManageMessage(args ...any) error // alias for Message
func (c *Context) LastPurged() int

type MessageAction int

const (
	MsgPin MessageAction = iota + 1
	MsgUnpin
	MsgDelete
	MsgCrosspost
	MsgReact
	MsgUnreact
	MsgClearReactions
	MsgEdit
	MsgPurge
)

// typing and channel shortcuts
func (c *Context) Typing() error
```

### Fluent Action Suites
```go
// member suite
func (c *Context) MemberSuite(targetID ...types.Snowflake) MemberSuite
func (m MemberSuite) ID() types.Snowflake
func (m MemberSuite) Get() (*types.Member, error)
func (m MemberSuite) Kick(reason ...string) error
func (m MemberSuite) Ban(deleteHistory any, reason ...string) error
func (m MemberSuite) Unban(reason ...string) error
func (m MemberSuite) Timeout(duration any, reason ...string) error
func (m MemberSuite) RemoveTimeout(reason ...string) error
func (m MemberSuite) AddRole(roleID types.Snowflake, reason ...string) error
func (m MemberSuite) RemoveRole(roleID types.Snowflake, reason ...string) error
func (m MemberSuite) SetNick(nick string, reason ...string) error
func (m MemberSuite) ResetNick(reason ...string) error
func (m MemberSuite) Mute(reason ...string) error
func (m MemberSuite) Unmute(reason ...string) error
func (m MemberSuite) Deaf(reason ...string) error
func (m MemberSuite) Undeaf(reason ...string) error
func (m MemberSuite) Move(channelID types.Snowflake, reason ...string) error
func (m MemberSuite) Disconnect(reason ...string) error

// channel suite
func (c *Context) ChannelSuite(channelID ...types.Snowflake) ChannelSuite
func (a ChannelSuite) ID() types.Snowflake
func (a ChannelSuite) Get() (*types.Channel, error)
func (a ChannelSuite) Send(args ...any) (*types.Message, error)
func (a ChannelSuite) Typing() error
func (a ChannelSuite) SetName(name string, reason ...string) (*types.Channel, error)
func (a ChannelSuite) SetTopic(topic string, reason ...string) (*types.Channel, error)
func (a ChannelSuite) SetSlowmode(duration any, reason ...string) (*types.Channel, error)
func (a ChannelSuite) Edit(payload rest.ModifyChannelPayload, reason ...string) (*types.Channel, error)
func (a ChannelSuite) Create(payload rest.CreateChannelPayload, reason ...string) (*types.Channel, error)
func (a ChannelSuite) Delete(reason ...string) error
func (a ChannelSuite) Lock(reason ...string) error
func (a ChannelSuite) Unlock(reason ...string) error
func (a ChannelSuite) Purge(count int, reason ...string) (int, error)

// message suite
func (c *Context) MessageSuite(target ...types.Snowflake) MessageSuite
func (c *Context) TargetMessage() MessageSuite
func (a MessageSuite) ID() types.Snowflake
func (a MessageSuite) ChannelID() types.Snowflake
func (a MessageSuite) Get() (*types.Message, error)
func (a MessageSuite) Edit(payload rest.MessagePayload) (*types.Message, error)
func (a MessageSuite) SetContent(content string) (*types.Message, error)
func (a MessageSuite) SetEmbeds(embeds ...types.Embed) (*types.Message, error)
func (a MessageSuite) SetComponents(components ...types.Component) (*types.Message, error)
func (a MessageSuite) Delete(reason ...string) error
func (a MessageSuite) Purge(count int, reason ...string) (int, error)
func (a MessageSuite) Pin(reason ...string) error
func (a MessageSuite) Unpin(reason ...string) error
func (a MessageSuite) React(emoji string) error
func (a MessageSuite) RemoveReaction(emoji string, userID ...types.Snowflake) error

// role suite
func (c *Context) RoleSuite(roleID ...types.Snowflake) RoleSuite
func (a RoleSuite) ID() types.Snowflake
func (a RoleSuite) Get() (*types.Role, error)
func (a RoleSuite) Create(payload rest.RolePayload, reason ...string) (*types.Role, error)
func (a RoleSuite) Edit(payload rest.RolePayload, reason ...string) (*types.Role, error)
func (a RoleSuite) Delete(reason ...string) error
func (a RoleSuite) SetName(name string, reason ...string) (*types.Role, error)
func (a RoleSuite) SetColor(color int, reason ...string) (*types.Role, error)
func (a RoleSuite) SetPermissions(perms types.Permission, reason ...string) (*types.Role, error)
func (a RoleSuite) SetHoist(hoist bool, reason ...string) (*types.Role, error)
func (a RoleSuite) SetMentionable(mentionable bool, reason ...string) (*types.Role, error)
func (a RoleSuite) Assign(memberID types.Snowflake, reason ...string) error
func (a RoleSuite) Remove(memberID types.Snowflake, reason ...string) error
```

### Option Retrieval (`ctx.Opt`)
```go
func (c *Context) Opt(name string) OptionVal

func (v OptionVal) Exists() bool
func (v OptionVal) String(fallback ...string) string
func (v OptionVal) Int(fallback ...int) int
func (v OptionVal) Int64(fallback ...int64) int64
func (v OptionVal) Float64(fallback ...float64) float64
func (v OptionVal) Bool(fallback ...bool) bool
func (v OptionVal) Snowflake(fallback ...types.Snowflake) types.Snowflake
func (v OptionVal) User() *types.User
func (v OptionVal) Member() *types.Member
func (v OptionVal) Role() *types.Role
func (v OptionVal) Channel() *types.Channel
func (v OptionVal) Attachment() *types.Attachment
```

### Component & Modal Values
```go
func (c *Context) Values() []string
func (c *Context) Value() string
func (c *Context) Input(name string) string
```

### Target Resolution & Hierarchy
```go
func (c *Context) TargetID() types.Snowflake
func (c *Context) TargetUser() *types.User
func (c *Context) TargetMember() *types.Member
func (c *Context) TargetUserID() types.Snowflake
func (c *Context) ResolvedMessage() *types.Message
func (c *Context) FocusedOption() *types.ApplicationCommandOption
func (c *Context) FocusedValue() string
func (c *Context) FocusedName() string
func (c *Context) CanModerate(targetMemberID types.Snowflake) (bool, error)
func (c *Context) CanManageRole(roleID types.Snowflake) (bool, error)

func (u *User) AvatarURL(args ...any) string
func (m *Member) AvatarURL(args ...any) string

type AvatarOptions struct {
	Size    int
	Global  bool
	GuildID types.Snowflake
}
```

### Context Navigation & Lifecycle
```go
func (c *Context) Subcommand() string
func (c *Context) SubcommandGroup() string
func (c *Context) Context() context.Context
func (c *Context) Cancel()
func (c *Context) IsAcknowledged() bool
func (c *Context) IsDeferred() bool
func (c *Context) ResponseChannel() <-chan types.InteractionResponse
func (c *Context) Latency() time.Duration
func (c *Context) Set(key string, val any)
func (c *Context) Get(key string) any
```

## Components

### Embed Builder
```go
func NewEmbed() *EmbedBuilder
func (b *EmbedBuilder) SetTitle(title string) *EmbedBuilder
func (b *EmbedBuilder) SetDescription(desc string) *EmbedBuilder
func (b *EmbedBuilder) SetColor(color any) *EmbedBuilder
func (b *EmbedBuilder) SetURL(url string) *EmbedBuilder
func (b *EmbedBuilder) AddField(name, value string, inline ...bool) *EmbedBuilder
func (b *EmbedBuilder) AddFields(fields ...types.EmbedField) *EmbedBuilder
func (b *EmbedBuilder) ClearFields() *EmbedBuilder
func (b *EmbedBuilder) SetFooter(text string, iconURL ...string) *EmbedBuilder
func (b *EmbedBuilder) SetAuthor(name string, urls ...string) *EmbedBuilder
func (b *EmbedBuilder) SetThumbnail(url string) *EmbedBuilder
func (b *EmbedBuilder) SetImage(url string) *EmbedBuilder
func (b *EmbedBuilder) SetTimestamp(timestamp any) *EmbedBuilder
func (b *EmbedBuilder) ToEmbed() types.Embed
func (b *EmbedBuilder) Build() (*types.Embed, error)
```

### Buttons & Action Rows
```go
func Button(args ...any) types.Component

type ButtonOptions struct {
	Style    any
	CustomID string
	Label    string
	URL      string
	Emoji    any
	Disabled bool
}

func ActionRow(components ...types.Component) types.Component
func Row(components ...types.Component) types.Component
func PackRows(components ...types.Component) []types.Component
func ValidateRows(rows []types.Component) error
```

### Select Menus
```go
func StringSelect(customID, placeholder string, options ...types.SelectOption) types.Component
func UserSelect(customID, placeholder string, opts ...SelectOpt) types.Component
func RoleSelect(customID, placeholder string, opts ...SelectOpt) types.Component
func ChannelSelect(customID, placeholder string, opts ...SelectOpt) types.Component
func MentionableSelect(customID, placeholder string, opts ...SelectOpt) types.Component
func SelectOption(label, value string, description ...string) types.SelectOption
```

### Text Inputs
```go
func TextInputShort(customID, label string, opts ...TextInputOpt) types.Component
func TextInputParagraph(customID, label string, opts ...TextInputOpt) types.Component
func WithPlaceholder(placeholder string) TextInputOpt
func WithValue(value string) TextInputOpt
func WithMinLength(min int) TextInputOpt
func WithMaxLength(max int) TextInputOpt
func WithRequired(required bool) TextInputOpt
```

### Modals
```go
type Modal struct {
	CustomID   string
	Title      string
	Components []types.Component
}
```

## Middleware

```go
type CrashLogger interface {
	Crash(msg string, recovered any, stack []byte)
}

func Guard(opts ...GuardOption) Middleware
func WithPanicHandler(fn PanicHandler) GuardOption
func WithGuardLogger(l CrashLogger) GuardOption
func WithMessage(msg string) GuardOption
func WithEphemeral(ephemeral bool) GuardOption

func Cooldown(policies []Policy, opts ...CooldownOption) Middleware
func UserCooldown(duration any, burst ...int) Policy
func ChannelCooldown(duration any, burst ...int) Policy
func GuildCooldown(duration any, burst ...int) Policy
func GlobalCooldown(duration any, burst ...int) Policy

func RequireUser(perms types.Permission, opts ...PermissionOption) Middleware
func RequireBot(perms types.Permission, opts ...PermissionOption) Middleware
func RequirePermissions(userPerms, botPerms types.Permission, opts ...PermissionOption) Middleware
func Chain(middlewares ...Middleware) Middleware
```

## Utilities

```go
func NewLogger(configs ...LoggerConfig) (*FileLogger, error)
func ParseLogFormat(s string) LogFormat

func ParseDuration(s string) (time.Duration, error)
func FormatDuration(d time.Duration) string
func FormatDurationHuman(d time.Duration) string
func Time(input any) TimeItem

func (i TimeItem) Discord(style ...string) string
func (i TimeItem) Human() string
func (i TimeItem) Time() time.Time
func (i TimeItem) Duration() time.Duration
func (i TimeItem) Unix() int64
func (i TimeItem) UnixMilli() int64
func (i TimeItem) ISO() string

// Pointer helper exported under github.com/sfossl/gdk/util:
// func Ptr[T any](v T) *T
```

## Error Sentinels

All framework error sentinels are exported under `gdk.*` for comparison with `errors.Is`:

### Application & Lifecycle (`gdk/app`)
```go
var (
	ErrTokenRequired     // discord bot token is empty
	ErrAppIDRequired     // discord application id is empty
	ErrCommandNotFound   // interaction references an unregistered command
	ErrComponentNotFound // interaction references an unregistered component custom id
	ErrModalNotFound     // interaction references an unregistered modal custom id
	ErrAlreadyRunning    // App.Run called while application is already active
	ErrInvalidSignature  // ed25519 webhook request signature verification failed
	ErrPublicKeyRequired // ed25519 public key required in webhook mode
	ErrDuplicateCommand  // duplicate command name registered for the same type
	ErrDevGuildRequired  // dev mode enabled without providing dev guild id
	ErrNilModule         // nil module passed to RegisterModule or Module
)
```

### Command Schema Validation (`gdk/command`)
```go
var (
	ErrInvalidName                 // command or option name contains invalid characters
	ErrNameTooLong                 // command or option name exceeds 32 characters
	ErrEmptyName                   // command or option name is empty
	ErrEmptyDescription            // chat input command description is empty
	ErrDescriptionTooLong          // command description exceeds 100 characters
	ErrSubcommandTopOptions        // command mixes top-level options with subcommands
	ErrTooManyOptions              // options exceed discord maximum of 25
	ErrTooManyChoices              // option choices exceed discord maximum of 25
	ErrAutocompleteWithChoices     // option enables both autocomplete and static choices
	ErrEmptySubcommandGroup        // subcommand group defined with zero subcommands
	ErrEmptyChoiceName             // choice name is empty
	ErrChoiceNameTooLong           // choice name exceeds 100 characters
	ErrInvalidChoiceValue          // choice value type mismatch with option type
	ErrRequiredOptionAfterOptional // required option placed after optional option
	ErrDuplicateOptionName         // duplicate option name within the same level
	ErrDuplicateChoiceName         // duplicate choice name within the same option
	ErrContextMenuHasDescription   // context menu command has a description set
	ErrContextMenuHasOptions       // context menu command has options set
)
```

### Component & Embed Validation (`gdk/component`)
```go
var (
	ErrInvalidColor            // color value cannot be parsed
	ErrColorOutOfRange         // integer color exceeds 0xFFFFFF
	ErrEmbedTitleTooLong       // embed title exceeds 256 characters
	ErrEmbedDescriptionTooLong // embed description exceeds 4096 characters
	ErrEmbedTooManyFields      // embed contains more than 25 fields
	ErrEmbedFieldNameTooLong   // embed field name exceeds 256 characters
	ErrEmbedFieldValueTooLong  // embed field value exceeds 1024 characters
	ErrEmbedFieldMissingName   // embed field missing name
	ErrEmbedFieldMissingValue  // embed field missing value
	ErrEmbedFooterTooLong      // embed footer text exceeds 2048 characters
	ErrEmbedAuthorTooLong      // embed author name exceeds 256 characters
	ErrEmbedTotalLengthTooLong // total embed characters exceed 6000
	ErrTooManyActionRows       // message contains more than 5 action rows
	ErrRowTooManyButtons       // action row contains more than 5 buttons
	ErrRowMixedComponents      // action row mixes buttons with select menus or inputs
	ErrRowMultipleSelects      // action row contains more than 1 select menu
	ErrRowMultipleInputs       // action row contains more than 1 text input
	ErrEmptyActionRow          // action row contains zero components
	ErrNonActionRowAtRoot      // non-action-row component placed at message root
	ErrSelectMissingOptions    // string select menu contains zero options
	ErrSelectTooManyOptions    // string select menu contains more than 25 options
)
```

### Context & Entity Modules (`gdk/context`)
```go
var (
	ErrNotInGuild               // interaction invoked outside a guild
	ErrNoRESTClient             // rest client unavailable on context
	ErrCannotModerateSelf       // moderation attempted on the invoking user or bot
	ErrCannotModerateOwner      // moderation attempted on the guild owner
	ErrTargetOutranksUser       // target member outranks invoking user in role hierarchy
	ErrTargetOutranksBot        // target member outranks bot in role hierarchy
	ErrInvalidTarget            // target snowflake is 0 or invalid
	ErrInvalidModType           // unknown moderation action type
	ErrAlreadyResponded         // interaction callback already sent
	ErrModalAfterDefer          // modal callback cannot be sent after interaction deferral
	ErrNoInteraction            // context interaction payload is nil
	ErrEmptyMessage             // message dispatched with no content, embeds, components or files
	ErrEmptyMemberUpdate        // member update invoked with zero modified fields
	ErrTimeoutTooLong           // timeout duration exceeds discord maximum of 28 days
	ErrTimeoutTooShort          // timeout duration is negative or zero
	ErrBanDeleteTooLong         // ban message purge duration exceeds discord maximum 7 days
	ErrInvalidChannelID         // channel snowflake is 0 or invalid
	ErrChannelNameRequired      // channel name is empty on creation
	ErrSlowmodeTooLong          // slowmode duration exceeds discord maximum of 6 hours
	ErrSlowmodeNegative         // slowmode duration is negative
	ErrInvalidPurgeCount        // message purge count outside valid range (2-100)
	ErrMessagesTooOld           // messages older than 14 days cannot be bulk-purged
	ErrInvalidChannelAction     // unknown channel action type
	ErrInvalidRoleID            // role snowflake is 0 or invalid
	ErrRoleNotFound             // target role not found in guild
	ErrRoleOutranksBot          // target role outranks bot in role hierarchy
	ErrRoleOutranksUser         // target role outranks invoking user in role hierarchy
	ErrCannotModifyEveryoneRole // modification of @everyone role is forbidden
	ErrInvalidRoleAction        // unknown role action type
	ErrInvalidMessageID         // message snowflake is 0 or invalid
	ErrInvalidMessageAction     // unknown message action type
	ErrEmojiRequired            // emoji parameter is empty on reaction action
)
```

### Middleware (`gdk/middleware`)
```go
var (
	ErrPanicRecovered      // panic caught and recovered by Guard middleware
	ErrOnCooldown          // command rate limit bucket is currently exhausted
	ErrInvalidCooldown     // invalid cooldown duration or burst configuration
	ErrMissingGuildInfo    // guild id or permissions missing from interaction payload
	ErrMissingPermissions  // caller or bot lacks required permissions
	ErrNoMemberPermissions // member permissions unavailable on interaction
	ErrNoBotPermissions    // bot application permissions unavailable on interaction
)
```

### Utilities (`gdk/util`)
```go
var (
	ErrInvalidDuration // duration string cannot be parsed
	ErrEmptyDuration   // duration string is empty
	ErrUnknownUnit     // duration contains unknown time unit
	ErrInvalidTime     // invalid time argument passed to Time
	ErrUnsupportedType // unsupported type passed to utility conversion
)
```

---

## Connected Topics

| Context | Destination | Description |
| :--- | :--- | :--- |
| **Index** | [Docs Overview](../README.md#documentation) | Main documentation directory and package architecture |
| **Getting Started** | [Getting Started](getting-started.md) | Bootstrap, lifecycle, configuration and environment setup |
| **Commands** | [Commands](commands.md) | Slash commands, options, choices and routing |
| **Context** | [Context](context.md) | Execution context, option extraction and responses |
| **Entity Modules** | [Entity Modules](modules/index.md) | Unified modules for Respond, Member, Channel, Role, Message |
| **Components** | [Components & Modals](components.md) | Embed builders, buttons, select menus and modal dialogs |
| **Middleware** | [Middleware](middleware.md) | Token-bucket cooldowns, permission guards and panic recovery |
| **Sync** | [Synchronization](sync.md) | Differential command hashing and schema synchronization |
| **Transports** | [Transports](transports.md) | Gateway WebSocket and HTTP Webhook execution models |
| **Utilities** | [Utilities](utilities.md) | Time formatting, duration parsing, pointer helper and logger |
