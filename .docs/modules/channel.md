[Docs Overview](../../README.md#documentation) / [Context](../context.md) / [Entity Modules](index.md) / Channel

# Channel Module

`ctx.Channel` is the channel management engine in `gdk`. It provides direct action execution and fluent operations for creating, editing, deleting, locking, unlocking and managing Discord channels.

---

## Method Signatures

```go
// direct action executor (options or positional shortcuts)
func (c *Context) Channel(args ...any) error
func (c *Context) ManageChannel(args ...any) error // backwards-compatible alias

// fluent suite for channel inspection, queries and chained updates
func (c *Context) ChannelSuite(channelID ...types.Snowflake) ChannelSuite

// typing indicator
func (c *Context) Typing() error
```

`Channel` accepts universal `gdk.Options`, dedicated type-safe `gdk.ChannelOptions` or convenient positional shortcuts.

---

## Typed `ChannelOptions` Struct

For compile-time type safety without generic field resolution, pass `gdk.ChannelOptions`:

```go
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

// usage example
err := ctx.Channel(gdk.ChannelOptions{
	Type:     gdk.ChannelCreate,
	Name:     "channel-name",
	Topic:    "channel topic",
	Slowmode: "15s",
	Reason:   "reason",
})
```

---

## Options Field Reference

When calling `ctx.Channel(gdk.Options{...})`, the following fields are processed:

| Field | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `Type` | `any` | `nil` | action type (`"create"`, `"edit"`, `"delete"`, `"lock"`, `"unlock"`, `"purge"`) |
| `Target` / `Channel` | `gdk.Snowflake` | `ctx.ChannelID` | channel snowflake to operate on |
| `Name` | `string` | `""` | channel name (required when creating channels) |
| `Topic` | `string` | `""` | channel description / topic |
| `Slowmode` | `any` | `nil` | rate limit per user (seconds `int`, `time.Duration` or string `"5s"`, `"10m"`) |
| `Count` | `int` | `0` | message count (2-100) when forwarding purge actions |
| `Reason` | `string` | `""` | audit log reason sent in `X-Audit-Log-Reason` |

---

## Positional Shortcuts

```go
// create channel in current guild
ctx.Channel("create", "channel-name", "reason")

// create channel with flexible argument ordering (name first)
ctx.Channel("channel-name", gdk.ChannelCreate)

// edit channel name or topic
ctx.Channel(channelID, "edit", "new-channel-name")

// lock channel (deny @everyone SendMessages)
ctx.Channel("lock", "reason")

// unlock channel (remove @everyone SendMessages overwrite)
ctx.Channel("unlock", "reason")

// delete channel with audit reason
ctx.Channel(channelID, "delete", "reason")
```

---

## Structured Options Examples

### Creating a Channel

```go
ctx.Channel(gdk.Options{
    Type:   "create",
    Name:   "channel-name",
    Topic:  "channel topic",
    Reason: "reason",
})
```

### Locking Down a Channel

```go
ctx.Channel(gdk.Options{
    Target: channelID,
    Type:   "lock",
    Reason: "reason",
})
```

### Setting Slowmode with Human Duration

```go
ctx.Channel(gdk.Options{
    Target:   channelID,
    Type:     "edit",
    Slowmode: "30s",
    Reason:   "reason",
})
```

---

## Companion Suite (`ctx.ChannelSuite`)

For chained queries, message sending, typing triggers and granular edits:

```go
suite := ctx.ChannelSuite(channelID)

// fetch full channel metadata
ch, err := suite.Get()

// send chat message directly into channel
msg, err := suite.Send("message content")

// trigger typing indicator
err := suite.Typing()

// fluent mutations
suite.SetName("channel-name", "reason")
suite.SetTopic("channel topic")
suite.SetSlowmode("10s")
suite.Lock("reason")
suite.Unlock("reason")
suite.Delete("reason")
deleted, err := suite.Purge(50, "reason")
```

---

## Typed Errors

| Error Sentinel | Condition |
| :--- | :--- |
| `gdk.ErrInvalidChannelID` | Channel ID is zero or unresolvable |
| `gdk.ErrChannelNameRequired` | Channel name is empty when creating a channel |
| `gdk.ErrSlowmodeTooLong` | Slowmode duration exceeds Discord's maximum 6 hours |
| `gdk.ErrSlowmodeNegative` | Slowmode duration is negative |
| `gdk.ErrInvalidChannelAction` | Unrecognized action type passed to `ctx.Channel` |
| `gdk.ErrNotInGuild` | Attempting guild-scoped channel actions outside a guild |
| `gdk.ErrNoRESTClient` | REST client is uninitialized on gateway context |

---

## Connected Topics

| Context | Destination | Description |
| :--- | :--- | :--- |
| **Modules Hub** | [Entity Modules Overview](index.md) | Universal options matrix and engine architecture |
| **Context** | [Context Guide](../context.md) | Interaction lifecycle, channel resolving and deferral |
| **Utilities** | [Time Utilities](../utilities.md#human-duration-parsing) | Human duration parsing for channel slowmode durations |
| **Respond** | [respond.md](respond.md) | Replies, direct messages and standalone channel messages |
| **Message** | [message.md](message.md) | Bulk channel purges, pins and reaction management |
| **Member** | [member.md](member.md) | Voice channels, member moving and guild moderation |
| **Role** | [role.md](role.md) | Channel permission overrides and role assignments |
| **API Reference** | [ChannelSuite](../api-reference.md#fluent-action-suites) | Fluent companion suite and error sentinels |

