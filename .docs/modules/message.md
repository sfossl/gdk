[Docs Overview](../../README.md#documentation) / [Context](../context.md) / [Entity Modules](index.md) / Message

# Message Module

`ctx.Message` is the chat message operation engine in `gdk`. It handles standalone message actions: pinning, unpinning, deleting, crossposting announcements, reactions (unicode, custom static and animated server emojis) and editing older messages by their Snowflake ID.

---

## Method Signatures

```go
// direct action executor (options or positional shortcuts)
func (c *Context) Message(args ...any) error

// fluent suite for queries, edits and chained operations
func (c *Context) MessageSuite(messageID types.Snowflake, channelID ...types.Snowflake) MessageSuite
```

`Message` accepts universal `gdk.Options`, dedicated type-safe `gdk.MessageOptions` or concise positional shortcuts.

---

## Typed `MessageOptions` Struct

For compile-time type safety without generic field resolution, pass `gdk.MessageOptions`:

```go
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

// usage example
err := ctx.Message(gdk.MessageOptions{
	Type:      gdk.MsgPin,
	MessageID: targetMsgID,
	Reason:    "reason",
})
```

---

## Options Field Reference

When calling `ctx.Message(gdk.Options{...})`, the following fields are processed:

| Field | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `Target` | `gdk.Snowflake` | `ctx.TargetID()` | target message snowflake (defaults to clicked message in context menus) |
| `Channel` | `gdk.Snowflake` | `ctx.ChannelID` | channel containing the target message |
| `Type` | `any` | `nil` | action type (`"pin"`, `"unpin"`, `"delete"`, `"crosspost"`, `"react"`, `"unreact"`, `"clearreactions"`, `"edit"`, `"purge"`) |
| `Reason` | `string` | `""` | audit log reason for pins, deletions and purges |
| `Count` | `int` | `0` | number of messages to delete (2-100) when `Type` is `"purge"` |
| `Emoji` | `any` | `nil` | single emoji string or `[]string` slice for reactions |
| `Content` | `string` | `""` | updated message text body when `Type` is `"edit"` |
| `Embeds` | `[]gdk.Embed` | `nil` | updated embeds when `Type` is `"edit"` |
| `Components` | `[]gdk.Component` | `nil` | updated action rows when `Type` is `"edit"` |
| `Files` | `[]types.File` | `nil` | file attachments when `Type` is `"edit"` |

---

## Emoji Support & Auto-Sanitization

`ctx.Message` accepts all Discord emoji formats and automatically normalizes them:

- **Unicode Emojis**: `"🦆"`
- **Custom Static Emojis**: `"custom_emoji:1234567890"` or chat mention `"<:custom_emoji:1234567890>"`
- **Custom Animated Emojis**: `"custom_animated:1234567890"` or chat mention `"<a:custom_animated:1234567890>"`

Angle brackets (`<:`, `<a:`, `>`) are automatically stripped and path-escaped, so raw chat tags paste directly without error.

---

## Positional Shortcuts

```go
// pin message
ctx.Message(messageID, "pin", "reason")

// unpin message
ctx.Message(messageID, "unpin")

// delete message
ctx.Message(messageID, "delete", "reason")

// publish announcement to follower channels
ctx.Message(messageID, "crosspost")

// add reaction
ctx.Message(messageID, "react", "🦆")

// add custom server emoji
ctx.Message(messageID, "react", "<:custom_emoji:123456789>")

// remove bot's reaction
ctx.Message(messageID, "unreact", "🦆")

// clear all reactions from message
ctx.Message(messageID, "clearreactions")

// clear all reactions for a specific emoji
ctx.Message(messageID, "clearreactions", "🦆")

// edit older message text
ctx.Message(messageID, "edit", "updated message content")

// bulk purge messages in current channel
ctx.Message("purge", 50, "reason")
```

---

## Structured Options Examples

### Adding Reactions

```go
ctx.Message(gdk.Options{
    Target: messageID,
    Type:   "react",
    Emoji:  []string{"🦆", "<:custom_emoji:123456789>"},
})
```

### Pinning a Message with Audit Reason

```go
ctx.Message(gdk.Options{
    Target: messageID,
    Type:   "pin",
    Reason: "reason",
})
```

### Deleting a Message in Another Channel

```go
ctx.Message(gdk.Options{
    Target:  messageID,
    Channel: channelID,
    Type:    "delete",
    Reason:  "reason",
})
```

### Editing an Older Bot Message by ID

```go
ctx.Message(gdk.Options{
    Target:     messageID,
    Type:       "edit",
    Content:    "updated message content",
    Embeds:     []gdk.Embed{embed},
    Components: []gdk.Component{row},
})
```

---

## Bulk Message Purge

`ctx.Message` provides bulk deletion with automatic 14-day filtering and actual count tracking:

- **Count constraints**: Must be between 2 and 100 (`gdk.ErrInvalidPurgeCount` returned if outside bounds).
- **14-day filter**: Discord rejects bulk deletion of messages older than 14 days. `gdk` automatically filters out messages exceeding this boundary.
- **Old message sentinel**: If all fetched messages are older than 14 days, `ctx.Message` returns `gdk.ErrMessagesTooOld`.
- **Actual count resolution**: Query `ctx.LastPurged()` to read the exact count of deleted messages or call `ctx.MessageSuite().Purge(count)` which returns `(int, error)`.

```go
if err := ctx.Message("purge", 50, "bulk chat cleanup"); err != nil {
	if errors.Is(err, gdk.ErrMessagesTooOld) {
		return ctx.Respond("no messages could be purged because all messages are older than 14 days (discord limit).", true)
	}
	return ctx.Respond("failed to purge messages: "+err.Error(), true)
}

deleted := ctx.LastPurged()
return ctx.Respond(gdk.Options{
	Content:     fmt.Sprintf("successfully purged %d messages.", deleted),
	DeleteAfter: 5 * time.Second,
})
```

---

## Companion Suite (`ctx.MessageSuite`)

For chained operations or querying message data:

```go
suite := ctx.MessageSuite(messageID)

// fetch full message metadata
msg, err := suite.Get()

// fluent mutations
suite.Pin("reason")
suite.Unpin()
suite.Delete("reason")
suite.Crosspost()
suite.React("🦆")
suite.Unreact("🦆")
suite.ClearReactions()

// fluent edits
suite.SetContent("updated message content")
suite.SetEmbeds(embed)
suite.SetComponents(row)
```

---

## Typed Errors

| Error Sentinel | Condition |
| :--- | :--- |
| `gdk.ErrInvalidMessageID` | Target message ID is zero or unresolved |
| `gdk.ErrInvalidChannelID` | Target channel ID is zero or unresolved |
| `gdk.ErrNoRESTClient` | REST client is uninitialized on gateway context |
| `gdk.ErrEmojiRequired` | Reaction action called with empty emoji |
| `gdk.ErrInvalidPurgeCount` | Purge count is outside allowed 2-100 bounds |
| `gdk.ErrMessagesTooOld` | All targeted messages are older than 14 days and cannot be bulk deleted |
| `gdk.ErrInvalidMessageAction` | Message action type is unrecognized |

---

## Connected Topics

| Context | Destination | Description |
| :--- | :--- | :--- |
| **Modules Hub** | [Entity Modules Overview](index.md) | Universal options matrix and engine architecture |
| **Context** | [Context Guide](../context.md) | Interaction lifecycle, target resolution and deferral |
| **Components** | [Components & Embeds](../components.md) | Embed builders, buttons, select menus and action rows |
| **Channel** | [channel.md](channel.md) | Channel creation, lock/unlock and slowmode controls |
| **Respond** | [respond.md](respond.md) | Interaction replies, followups and standalone messages |
| **Member** | [member.md](member.md) | Member moderation, timeouts and punishments |
| **Role** | [role.md](role.md) | Role creation, hierarchy validation and assignments |
| **API Reference** | [MessageSuite](../api-reference.md#fluent-action-suites) | Fluent companion suite and error sentinels |

