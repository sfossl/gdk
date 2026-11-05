[Docs Overview](../README.md#documentation) / Commands

# Commands

`gdk` provides an application command engine supporting Chat Input (slash commands), User context menus, Message context menus, built-in schema validation, option extraction, subcommands and autocomplete routing.

## Defining a Command

Commands are declared using `gdk.Command` and registered with a handler:

```go
cmd := gdk.Command{
	Name:        "info",
	Description: "display information",
	GuildOnly:   true,
	Permissions: gdk.PermissionSendMessages | gdk.PermissionViewChannel,
	Cooldown:    5 * time.Second,
	Options: []gdk.Option{
		gdk.StringOption("query", "search query").WithRequired(),
	},
	Execute: func(ctx *gdk.Context) error {
		query := ctx.Opt("query").String()
		_, err := ctx.Respond(fmt.Sprintf("result: %s", query))
		return err
	},
}

app.Command(cmd)
```

## Context Menu Commands

`gdk` natively supports Discord User and Message context menu commands (opened by right-clicking a user or message in the client). Discord requires that context menu commands omit descriptions and options:

### User Context Menu
```go
userMenu := gdk.Command{
	Type: gdk.CommandTypeUser,
	Name: "User Profile",
	Execute: func(ctx *gdk.Context) error {
		targetID := ctx.TargetID()
		user := ctx.TargetUser()

		return ctx.Respond(gdk.Options{
			Content:   fmt.Sprintf("User: %s (%s)", user.Username, targetID),
			Ephemeral: true,
		})
	},
}

app.Command(userMenu)
```

### Message Context Menu
```go
msgMenu := gdk.Command{
	Type: gdk.CommandTypeMessage,
	Name: "Pin Message",
	Execute: func(ctx *gdk.Context) error {
		if err := ctx.Message("pin"); err != nil {
			return err
		}
		_, err := ctx.Respond("Pinned message", true)
		return err
	},
}

app.Command(msgMenu)
```

## Option Types

`gdk` supports all 9 Discord application command option types:

| Option Constructor | Underlying Type | Return Extraction |
| :--- | :--- | :--- |
| `gdk.StringOption(name, desc)` | String | `ctx.Opt("name").String()` |
| `gdk.IntOption(name, desc)` | Integer (int64) | `ctx.Opt("name").Int()`, `.Int64()` |
| `gdk.NumberOption(name, desc)` | Floating point (float64) | `ctx.Opt("name").Float64()` |
| `gdk.BoolOption(name, desc)` | Boolean | `ctx.Opt("name").Bool()` |
| `gdk.UserOption(name, desc)` | User snowflake | `ctx.Opt("name").User()`, `ctx.Opt("name").Member()` |
| `gdk.ChannelOption(name, desc)` | Channel snowflake | `ctx.Opt("name").Channel()` |
| `gdk.RoleOption(name, desc)` | Role snowflake | `ctx.Opt("name").Role()` |
| `gdk.MentionableOption(name, desc)` | User or role snowflake | `ctx.Opt("name").Snowflake()` |
| `gdk.AttachmentOption(name, desc)` | Attachment snowflake | `ctx.Opt("name").Attachment()` |

## Option Modifiers

Options accept chained modifier methods:

```go
opt := gdk.StringOption("mode", "operating mode").
	WithRequired().
	WithChoice("fast", "fast").
	WithChoice("slow", "slow").
	WithLength(3, 20)

intOpt := gdk.IntOption("count", "total count").
	WithMinMax(1, 100)

chanOpt := gdk.ChannelOption("target", "announcement channel", 0, 5)
```

### Pointer Constraints (`util.Ptr`)

Optional command and option constraints (`MinLength`, `MaxLength`, `MinValue`, `MaxValue`, `DMPermission`) require pointer types so that `nil` denotes an unset constraint. Because Go does not allow taking the memory address of constant literals directly (e.g. `&10` causes a compiler error), `gdk/util` provides `util.Ptr[T any](v T) *T`:

```go
import "github.com/sfossl/gdk/util"

opt := gdk.Option{
	Type:        gdk.OptionTypeString,
	Name:        "username",
	Description: "target username",
	MinLength:   util.Ptr(3),
	MaxLength:   util.Ptr(32),
}

cmd := gdk.Command{
	Name:         "secret",
	Description:  "guild-only command",
	DMPermission: util.Ptr(false),
}
```

## Permissions and Enforcement

`gdk.Command` provides two complementary permission fields:

| Field | Target | Evaluated By | Behavior |
| :--- | :--- | :--- | :--- |
| `DefaultMemberPermissions` | Client UI Visibility | Discord API | Controls whether the command appears visible in the Discord client slash menu. Pass a pointer to a `types.Permission` bitmask. |
| `Permissions` | Runtime Server Guard | `gdk` Engine | Hard server-side enforcement. When non-zero, `gdk` automatically prepends `middleware.RequireUser(cmd.Permissions)` to the command pipeline ahead of `Cooldown`. |

Because `Permissions` runs ahead of `Cooldown` in the pipeline, unauthorized members without the required permissions fail immediately and cannot drain cooldown tokens for authorized users:

```go
adminCmd := gdk.Command{
	Name:        "purge",
	Description: "purge channel messages",
	Permissions: gdk.PermissionManageMessages,
	Cooldown:    10 * time.Second,
	Execute: func(ctx *gdk.Context) error {
		return ctx.Respond("purged")
	},
}
```

## Command Cooldowns

`gdk.Command` provides built-in token-bucket rate limiting via two struct fields:

| Field | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `Cooldown` | `time.Duration` | `0` (disabled) | Duration of the rate limit bucket (e.g. `10 * time.Second`). |
| `CooldownScope` | `CooldownScope` | `CooldownScopeUser` | Target boundary: `CooldownScopeUser`, `CooldownScopeChannel`, `CooldownScopeGuild`, `CooldownScopeGlobal`. |

> [!IMPORTANT]
> `Cooldown` requires a typed `time.Duration` (e.g. `10 * time.Second`, `24 * time.Hour`).
> - Passing a string literal like `"10s"` fails compilation because Go is statically typed.
> - Passing an untyped integer like `10` sets `10ns` (nanoseconds), which expires immediately.
> - To dynamically parse human strings inside a struct literal, use `gdk.Time("10s").Duration()`.
> - If you need flexible input types (`"10s"`, `10`) or burst capacities, use the policy helpers `gdk.UserCooldown(...)` in middleware.

## Dynamic Autocomplete

For options that fetch live choices on user keystrokes, enable `.WithAutocomplete()` and provide an `AutocompleteHandler`. You can use `ctx.FocusedOption()` to inspect which option the user is currently typing into (works at top-level or within nested subcommands):

```go
cmd := gdk.Command{
	Name:        "tag",
	Description: "lookup a stored tag",
	Options: []gdk.Option{
		gdk.StringOption("name", "tag identifier").WithAutocomplete().WithRequired(),
	},
	AutocompleteHandler: func(ctx *gdk.Context) error {
		query := ctx.FocusedValue()
		_ = query

		return ctx.Autocomplete([]gdk.Choice{
			{Name: "one", Value: "1"},
			{Name: "two", Value: "2"},
		})
	},
	Execute: func(ctx *gdk.Context) error {
		tag := ctx.Opt("name").String()
		_, err := ctx.Respond(fmt.Sprintf("tag: %s", tag))
		return err
	},
}

app.Command(cmd)
```

## Subcommands and Groups

Commands can branch into nested subcommands. Handlers can be attached **directly to subcommands** or routed via `switch ctx.Subcommand()` in the top-level handler:

### Direct Subcommand Handlers
```go
cmd := gdk.Command{
	Name:        "config",
	Description: "manage configuration",
	Subcommands: []gdk.Subcommand{
		{
			Name:        "get",
			Description: "read configuration value",
			Options: []gdk.Option{
				gdk.StringOption("key", "setting key").WithRequired(),
			},
			Execute: func(ctx *gdk.Context) error {
				key := ctx.Opt("key").String()
				_, err := ctx.Respond(fmt.Sprintf("key: %s", key))
				return err
			},
		},
		{
			Name:        "set",
			Description: "write configuration value",
			Options: []gdk.Option{
				gdk.StringOption("key", "setting key").WithRequired(),
				gdk.StringOption("val", "setting value").WithRequired(),
			},
			Execute: func(ctx *gdk.Context) error {
				_, err := ctx.Respond("saved")
				return err
			},
		},
	},
}

app.Command(cmd)
```

### Top-Level Command Router
Alternatively, you can register a top-level handler and inspect active branches using `ctx.Subcommand()`:

```go
cmd := gdk.Command{
	Name:        "config",
	Description: "manage configuration",
	Subcommands: []gdk.Subcommand{
		{
			Name:        "get",
			Description: "read configuration value",
			Options: []gdk.Option{
				gdk.StringOption("key", "setting key").WithRequired(),
			},
		},
		{
			Name:        "set",
			Description: "write configuration value",
			Options: []gdk.Option{
				gdk.StringOption("key", "setting key").WithRequired(),
				gdk.StringOption("val", "setting value").WithRequired(),
			},
		},
	},
	Execute: func(ctx *gdk.Context) error {
		switch ctx.Subcommand() {
		case "get":
			key := ctx.Opt("key").String()
			_, err := ctx.Respond(fmt.Sprintf("key: %s", key))
			return err
		case "set":
			_, err := ctx.Respond("saved")
			return err
		default:
			return nil
		}
	},
}

app.Command(cmd)
```

### Subcommand Inspection
Inspect active branches inside a top-level handler:

```go
group := ctx.SubcommandGroup()
sub := ctx.Subcommand()
```

## Global Self-Registration

Commands can self-register globally in module `init()` functions without manual wiring in `main.go`:

```go
func init() {
	gdk.Register(Info)
}

var Info = gdk.Command{
	Name:        "info",
	Description: "display info",
	Execute:     info,
}

func info(ctx *gdk.Context) error {
	_, err := ctx.Respond("pong")
	return err
}
```

During startup, call `app.MountRegistered()` to register all mounted commands automatically.

## Client Validation

`gdk.ValidateCommand(cmd)` verifies schemas locally before syncing to Discord:
- Slash command names must be 1 to 32 characters matching `^[-_\p{L}\p{N}\p{sc=Deva}\p{sc=Thai}]+$`.
- Context menu command names must be 1 to 32 characters (can include uppercase and spaces).
- Slash command descriptions must be 1 to 100 characters; Context menu descriptions must be empty.
- Top-level commands can contain at most 25 options.
- Required options must precede optional options.
- Choices can have at most 25 entries per option.

---

## Connected Topics

| Context | Destination | Description |
| :--- | :--- | :--- |
| **Previous** | [Getting Started](getting-started.md) | Setup, configuration and app initialization |
| **Next** | [Context](context.md) | Handler context, option getters, response lifecycle and entity engines |
| **Components** | [Components & Modals](components.md) | Interactive buttons, select menus, action rows and modal dialogs |
| **Middleware** | [Middleware](middleware.md) | Token-bucket rate limits, permissions and panic guards |
| **Sync** | [Synchronization](sync.md) | Differential command hashing and schema synchronization |
| **API Reference** | [Commands API](api-reference.md#commands) | Command definitions, option constructors and schema validators |
