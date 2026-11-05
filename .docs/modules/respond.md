[Docs Overview](../../README.md#documentation) / [Context](../context.md) / [Entity Modules](index.md) / Respond

# Respond Module

`ctx.Respond` is the primary response engine in `gdk`. It routes immediate interaction callbacks, edits after deferrals, followup messages, direct messages with automatic receipts, author ping controls and standalone channel dispatches.

---

## Method Signature

```go
func (c *Context) Respond(args ...any) (*types.Message, error)
```

`Respond` accepts any combination of positional shortcuts (`string`, `bool`, `error`, `Embed`, `Component`, `File`) or a structured `gdk.Options` struct.

---

## Options Field Reference

When calling `ctx.Respond(gdk.Options{...})`, the following fields are processed:

| Field | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `Content` | `string` | `""` | message body text |
| `Ephemeral` | `bool` | `false` | send message visible only to the invoking user |
| `Deferred` | `bool` | `false` | triggers deferral instead of immediate response |
| `Scope` | `any` | `"guild"` | set to `"dm"` to route message to direct messages |
| `Target` | `gdk.Snowflake` | `ctx.TargetID()` $\to$ `ctx.Author.ID` | recipient user id when sending a dm |
| `Type` | `any` | `"reply"` | set to `"send"` to post a standalone channel message |
| `Channel` | `gdk.Snowflake` | `ctx.ChannelID` | target channel id for standalone channel messages |
| `Mention` | `*bool` | `nil` (discord default) | toggles replied author ping (`@ on` / `@ off`) |
| `Typing` | `bool` | `false` | triggers channel typing indicator before dispatch |
| `Embed` | `any` | `nil` | single rich embed object shortcut |
| `Embeds` | `[]gdk.Embed` | `nil` | rich embed objects (up to 10 per message) |
| `Components` | `[]gdk.Component` | `nil` | action rows with buttons or select menus (up to 5 rows) |
| `ComponentTimeout` | `any` | `0` | duration after which all buttons and select menus are disabled |
| `DeleteAfter` | `any` | `0` | duration after which the response message is automatically deleted |
| `Files` | `[]gdk.File` | `nil` | file attachments and uploads |
| `TTS` | `bool` | `false` | text-to-speech message flag |

---

## Pre-Flight Fast Failure

`ctx.Respond` validates before touching the network. If `Content`, `Embeds`, `Components` and `Files` are all empty (and `Deferred` is `false`), it returns `gdk.ErrEmptyMessage` immediately without sending a network request.

---

## Positional Shortcuts

```go
// plain text reply
ctx.Respond("pong")

// private ephemeral reply
ctx.Respond("only you see this", true)

// direct error reply
ctx.Respond(err)

// single embed
ctx.Respond(myEmbed)

// multiple embeds with ephemeral visibility
ctx.Respond(embed1, embed2, true)

// embed and button row
ctx.Respond(myEmbed, actionRow)

// text with file attachment
ctx.Respond("generated report", myFile)

// text merged with structured options
ctx.Respond("generated report", gdk.Options{Ephemeral: true})
```

---

## Structured Options Examples

### Standard Public Reply

```go
ctx.Respond(gdk.Options{
    Content: "response content",
})
```

### Ephemeral Reply with Embeds and Buttons

```go
ctx.Respond(gdk.Options{
    Content:    "response content",
    Ephemeral:  true,
    Embeds:     []gdk.Embed{embed},
    Components: []gdk.Component{row},
})
```

### Direct Message with Automatic Receipt

Opens the user's direct messages channel, sends the message and acknowledges the guild interaction ephemerally so the slash command does not time out.

```go
ctx.Respond(gdk.Options{
    Scope:   "dm",
    Content: "direct message content",
})
```

### Direct Message to Specific User

```go
ctx.Respond(gdk.Options{
    Scope:   "dm",
    Target:  userID,
    Content: "direct message content",
})
```

### Author Ping Suppression (`@ off`)

Replies to the command author without triggering a mention notification badge in Discord.

```go
noPing := false
ctx.Respond(gdk.Options{
    Content: "response without mention",
    Mention: &noPing,
})
```

### Standalone Channel Message Dispatch

Sends an independent message directly to a channel instead of linking it as a reply to the interaction.

```go
// post to specific channel
ctx.Respond(gdk.Options{
    Channel: channelID,
    Content: "message content",
})

// post to current channel as standalone message
ctx.Respond(gdk.Options{
    Type:    "send",
    Content: "message content",
})
```

### Typing Indicator

Triggers a Discord typing indicator on the target channel before dispatching the response.

```go
ctx.Respond(gdk.Options{
    Typing:  true,
    Content: "processing data query",
})
```

### Direct Deferral

Equivalent to calling `ctx.Defer(true)`.

```go
ctx.Respond(gdk.Options{
    Deferred:  true,
    Ephemeral: true,
})
```

---

## Interaction Lifecycle Flow

`ctx.Respond` tracks interaction state automatically:

1. **Initial Response**: Sends the initial Discord interaction response (`CallbackTypeChannelMessageWithSource`).
2. **Deferred State**: If the interaction was deferred earlier (manually or via `AutoDefer`), `Respond` edits the deferred original message (`EditOriginalInteractionResponse`).
3. **Followup State**: If an initial response was already sent, subsequent calls to `ctx.Respond` automatically dispatch as followup messages (`CreateFollowupMessage`).

```
ctx.Respond(...)
 ├── Empty check ─────> Empty? ─> ErrEmptyMessage (fast fail before network)
 ├── Scope == "dm" ───> YES ────> CreateDM + SendMessage + Acknowledge Receipt
 ├── Type == "send" ──> YES ────> SendMessage directly to channel
 ├── Is Deferred? ────> YES ────> EditOriginalInteractionResponse (Webhook PATCH)
 ├── Is Responded? ───> YES ────> CreateFollowupMessage (Webhook POST)
 └── Initial ─────────> NO  ────> CreateInteractionResponse (Callback POST)
```

---

## Companion Suite Shortcuts

The context provides fluent shortcuts for channel messaging and typing:

```go
// send standalone message to any channel
ctx.ChannelSuite(channelID).Send("server alert")

// trigger typing indicator on current channel
ctx.Typing()

// trigger typing indicator on specific channel
ctx.ChannelSuite(channelID).Typing()
```

---

## Typed Errors

| Error Sentinel | Condition |
| :--- | :--- |
| `gdk.ErrEmptyMessage` | Content, Embeds, Components and Files are all empty on non-deferred call |
| `gdk.ErrAlreadyResponded` | Calling low-level `ctx.Defer` or `ctx.Pong` when already acknowledged |
| `gdk.ErrNoInteraction` | Interaction context is missing |
| `gdk.ErrNoRESTClient` | REST client is uninitialized on gateway context |
| `gdk.ErrInvalidTarget` | Scope is `"dm"` but target user id cannot be resolved |
| `gdk.ErrInvalidChannelID` | Standalone channel dispatch lacks valid channel id |

---

## Connected Topics

| Context | Destination | Description |
| :--- | :--- | :--- |
| **Modules Hub** | [Entity Modules Overview](index.md) | Universal options matrix and engine architecture |
| **Context** | [Context Guide](../context.md) | Interaction lifecycle, option getters and deferral |
| **Components** | [Components & Modals](../components.md) | Interactive buttons, select menus and action rows |
| **Member** | [member.md](member.md) | Moderation, profiles and voice states |
| **Channel** | [channel.md](channel.md) | Channel creation, edits, locks and purges |
| **Role** | [role.md](role.md) | Role creation, hierarchy validation and assignments |
| **Message** | [message.md](message.md) | Message pinning, reactions and older message edits |
| **API Reference** | [Unified Methods](../api-reference.md#the-6-unified-methods) | Method definitions and error sentinels |

