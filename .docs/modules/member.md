[Docs Overview](../../README.md#documentation) / [Context](../context.md) / [Entity Modules](index.md) / Member

# Member Module

`ctx.Member` is the unified member management and moderation engine in `gdk`. It consolidates all guild member operations (punishments such as bans, kicks, timeouts; profile attributes such as nicknames, roles; and voice state such as muting, deafening, moves) under an entity-driven API with built-in client-side role hierarchy validation.

---

## Method Signatures

```go
// direct action executor (options or positional shortcuts)
func (c *Context) Member(args ...any) error

// fluent suite for queries and chained operations
func (c *Context) MemberSuite(targetID ...types.Snowflake) MemberSuite
```

`Member` accepts universal `gdk.Options`, dedicated type-safe `gdk.MemberOptions` or convenient positional shortcuts.

---

## Typed `MemberOptions` Struct

For full compile-time type safety without generic field resolution, pass `gdk.MemberOptions`:

```go
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

// usage example
err := ctx.Member(gdk.MemberOptions{
	Type:     gdk.ModTimeout,
	Target:   targetID,
	Duration: 15 * time.Minute,
	Reason:   "reason",
})
```

---

## Options Field Reference

When calling `ctx.Member(gdk.Options{...})`, the following fields are processed:

| Field | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `Target` | `gdk.Snowflake` | `ctx.TargetID()` | target guild member or user snowflake |
| `Type` | `any` | inferred | moderation or profile action type (`"ban"`, `"kick"`, `"timeout"`, etc.) |
| `Reason` | `string` | `""` | audit log reason sent via `X-Audit-Log-Reason` header |
| `Duration` | `any` | `nil` | timeout duration or ban message purge duration (`time.Duration`, `"10m"`, `"1d"` or seconds) |
| `Nickname` | `*string` | `nil` | updated guild nickname (pass empty string or `"resetnick"` to reset) |
| `AddRoles` | `[]gdk.Snowflake` | `nil` | role IDs to assign to the member |
| `RemoveRoles` | `[]gdk.Snowflake` | `nil` | role IDs to remove from the member |
| `Role` | `gdk.Snowflake` | `0` | single role convenience target for role assignments or removals |
| `VoiceMute` | `*bool` | `nil` | guild voice mute state |
| `VoiceDeaf` | `*bool` | `nil` | guild voice deafen state |
| `VoiceChannel` | `*types.Snowflake` | `nil` | target voice channel ID (pass `0` or `"disconnect"` to disconnect) |

---

## Pre-Flight Hierarchy Safety & Fast Failure

`ctx.Member` automatically verifies permission safety **before** contacting the Discord REST API:

1. **Role Hierarchy Validation (`CanModerate`)**:
   - Blocks action if the target is the guild owner (`ErrCannotModerateOwner`).
   - Blocks action if the target is the invoking user or bot itself (`ErrCannotModerateSelf`).
   - Blocks action if the target's highest role outranks or equals the bot's highest role (`ErrTargetOutranksBot`).
   - Blocks action if the target's highest role outranks or equals the invoking user's highest role (`ErrTargetOutranksUser`).
2. **Context-Level Rate Limit Caching**:
   Guild and bot member data are cached during the interaction lifecycle, eliminating duplicate `GET /guilds` queries during batch operations.
3. **Empty Mutation Fast Failure**:
   Calling `ctx.Member` with a target but no action or profile fields returns `gdk.ErrEmptyMemberUpdate` immediately without making unnecessary HTTP requests.

---

## Positional Shortcuts

```go
// ban with reason and 1-day message purge
ctx.Member(targetUserID, "ban", "reason", "1d")

// quick kick with audit reason
ctx.Member(targetUserID, "kick", "reason")

// timeout for 15 minutes
ctx.Member(targetUserID, "timeout", 15*time.Minute, "reason")

// remove timeout immediately
ctx.Member(targetUserID, "untimeout")

// revoke guild ban
ctx.Member(targetUserID, "unban", "reason")

// update nickname
ctx.Member(targetUserID, "nick", "new-nickname")

// assign single role directly
ctx.Member(targetUserID, roleID)
```

---

## Structured Options Examples

### Banning a Member with Message Purge

```go
ctx.Member(gdk.Options{
    Target:   targetUserID,
    Type:     "ban",
    Reason:   "reason",
    Duration: "7d",
})
```

### Applying a Timeout

```go
ctx.Member(gdk.Options{
    Target:   targetUserID,
    Type:     "timeout",
    Duration: 30 * time.Minute,
    Reason:   "reason",
})
```

### Removing a Timeout

```go
ctx.Member(gdk.Options{
    Target: targetUserID,
    Type:   "untimeout",
    Reason: "reason",
})
```

### Updating Roles and Nickname Simultaneously

Modify multiple member attributes in a single atomic Discord REST request:

```go
newNick := "new-nickname"
ctx.Member(gdk.Options{
    Target:      targetUserID,
    Nickname:    &newNick,
    AddRoles:    []gdk.Snowflake{roleID},
    RemoveRoles: []gdk.Snowflake{oldRoleID},
    Reason:      "reason",
})
```

### Managing Guild Voice State

```go
mute := true
deaf := true

// server mute and deafen
ctx.Member(gdk.Options{
    Target:    targetUserID,
    VoiceMute: &mute,
    VoiceDeaf: &deaf,
    Reason:    "voice disruption",
})

// move to another voice channel
ctx.Member(gdk.Options{
    Target:       targetUserID,
    VoiceChannel: &afkVoiceChannelID,
})
```

---

## Execution Lifecycle Flow

```
ctx.Member(...)
 ├── Empty check ─────────> Empty fields? ─> ErrEmptyMemberUpdate
 ├── Mod Action? ─────────> YES ──> CanModerate? ─> NO  ─> ErrTargetOutranks...
 │                                               └> YES ─> Kick/Ban/Timeout REST
 ├── Role Modification? ──> YES ──> CanManageRole? ─> NO ─> ErrRoleOutranks...
 │                                                └> YES ─> Add/Remove Member Role
 └── Profile / Voice ─────> ModifyGuildMember (PATCH /guilds/{id}/members/{id})
```

---

## Companion Suite (`ctx.MemberSuite`)

For chained operations or querying member information:

```go
suite := ctx.MemberSuite(targetUserID)

// fetch full guild member object
member, err := suite.Get()

// fluent moderation actions
suite.Kick("reason")
suite.Ban("1d", "reason")
suite.Timeout(10*time.Minute, "reason")
suite.RemoveTimeout()
suite.Unban("reason")

// fluent profile & role operations
suite.SetNick("new-nickname")
suite.ResetNick()
suite.AddRole(roleID)
suite.RemoveRole(oldRoleID)

// fluent voice controls
suite.Mute()
suite.Unmute()
suite.Deaf()
suite.Undeaf()
suite.Move(loungeChannelID)
suite.Disconnect()
```

---

## Typed Errors

| Error Sentinel | Condition |
| :--- | :--- |
| `gdk.ErrNotInGuild` | Member operation executed outside of a guild context |
| `gdk.ErrNoRESTClient` | REST client is uninitialized on gateway context |
| `gdk.ErrInvalidTarget` | Target member ID is zero or cannot be resolved |
| `gdk.ErrEmptyMemberUpdate` | Called with no moderation action, nickname, role or voice mutation |
| `gdk.ErrCannotModerateOwner` | Target is the guild owner |
| `gdk.ErrCannotModerateSelf` | Target is the bot application or the invoking user |
| `gdk.ErrTargetOutranksBot` | Target's highest role is equal to or higher than the bot's highest role |
| `gdk.ErrTargetOutranksUser` | Target's highest role is equal to or higher than the invoking user's highest role |
| `gdk.ErrTimeoutTooShort` | Timeout duration is less than 1 second |
| `gdk.ErrTimeoutTooLong` | Timeout duration exceeds the Discord maximum of 28 days |
| `gdk.ErrBanDeleteTooLong` | Ban delete message duration exceeds the Discord maximum of 7 days |
| `gdk.ErrInvalidModType` | Action type string or enum is unrecognized |

---

## Connected Topics

| Context | Destination | Description |
| :--- | :--- | :--- |
| **Modules Hub** | [Entity Modules Overview](index.md) | Universal options matrix and engine architecture |
| **Context** | [Context Guide](../context.md) | Interaction lifecycle, target resolution and hierarchy checks |
| **Utilities** | [Time Utilities](../utilities.md#human-duration-parsing) | Compound human duration strings for bans and timeouts |
| **Respond** | [respond.md](respond.md) | Confirmations, ephemeral moderation receipts and DMs |
| **Role** | [role.md](role.md) | Role creation, hierarchy validation and assignments |
| **Channel** | [channel.md](channel.md) | Channel creation, edits, locks and purges |
| **Message** | [message.md](message.md) | Message pinning, reactions and older message edits |
| **API Reference** | [MemberSuite](../api-reference.md#fluent-action-suites) | Fluent companion suite and error sentinels |

