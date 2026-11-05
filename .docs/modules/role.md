[Docs Overview](../../README.md#documentation) / [Context](../context.md) / [Entity Modules](index.md) / Role

# Role Module

`ctx.Role` is the guild role engine in `gdk`. It provides direct action execution and fluent operations for creating, editing, deleting, assigning and reordering roles (with automatic role hierarchy validation and cached guild role lookups).

---

## Method Signatures

```go
// direct action executor (options or positional shortcuts)
func (c *Context) Role(args ...any) (*types.Role, error)
func (c *Context) ManageRole(args ...any) (*types.Role, error) // alias

// fluent suite for role queries, edits and assignments
func (c *Context) RoleSuite(roleID ...types.Snowflake) RoleSuite

// role hierarchy inspection
func (c *Context) CanManageRole(roleID types.Snowflake) (bool, error)
```

`Role` accepts universal `gdk.Options`, dedicated type-safe `gdk.RoleOptions` or convenient positional shortcuts.

---

## Typed `RoleOptions` Struct

For compile-time type safety without generic field resolution, pass `gdk.RoleOptions`:

```go
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

// usage example
role, err := ctx.Role(gdk.RoleOptions{
	Type:   gdk.RoleCreate,
	Name:   "role-name",
	Color:  0x3498DB,
	Reason: "reason",
})
```

---

## Hierarchy Engine & Caching

Every mutation (`edit`, `delete`, `assign`, `remove`) verifies that both the bot and the calling user outrank the target role in the guild hierarchy:

1. **Bot Check**: The bot's highest role must strictly outrank the target role position (unless the bot is guild owner). Returns `gdk.ErrRoleOutranksBot` otherwise.
2. **Invoker Check**: The invoking user's highest role must strictly outrank the target role position (unless the user is guild owner). Returns `gdk.ErrRoleOutranksUser` otherwise.
3. **Protected Roles**: The `@everyone` role cannot be modified or deleted. Returns `gdk.ErrCannotModifyEveryoneRole`.
4. **Context Caching**: Cached guild lookups (`getGuildCached`, `getRolesCached`, `getBotMemberCached`) ensure role hierarchy checks within a single command lifecycle never execute redundant REST requests.

---

## Options Field Reference

When calling `ctx.Role(gdk.Options{...})`, the following fields are processed:

| Field | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `Type` | `any` | `nil` | action type (`"create"`, `"edit"`, `"delete"`) |
| `Target` / `Role` | `gdk.Snowflake` | `0` | target role snowflake |
| `Name` | `string` | `""` | role name |
| `Color` | `int` | `0` | integer RGB color (e.g. `0x3498DB`) |
| `Permissions` | `*types.Permission` | `nil` | permission bitfield overwrite |
| `Hoist` | `*bool` | `nil` | display role separately in online member list |
| `Mentionable` | `*bool` | `nil` | allow anyone to `@mention` this role |
| `Reason` | `string` | `""` | audit log reason sent in `X-Audit-Log-Reason` |

---

## Positional Shortcuts

```go
// create role with name
role, err := ctx.Role("create", "role-name")

// create role with flexible argument ordering (name first)
role, err := ctx.Role("role-name", gdk.RoleCreate)

// create role with name and audit reason
role, err := ctx.Role("create", "role-name", "reason")

// delete role with audit reason
_, err := ctx.Role("delete", roleID, "reason")

// positional role targeting
_, err := ctx.Role(roleID, "delete")
```

---

## Structured Options Examples

### Creating a Role

```go
hoist := true
mentionable := true
role, err := ctx.Role(gdk.Options{
    Type:        "create",
    Name:        "role-name",
    Color:       0xE67E22,
    Hoist:       &hoist,
    Mentionable: &mentionable,
    Reason:      "reason",
})
```

### Deleting a Role Safely

```go
_, err := ctx.Role(gdk.Options{
    Target: roleID,
    Type:   "delete",
    Reason: "reason",
})
```

---

## Companion Suite (`ctx.RoleSuite`)

For chained mutations, role assignment/removal and queries:

```go
suite := ctx.RoleSuite(roleID)

// fetch role details from guild cache/REST
role, err := suite.Get()

// assign or revoke from a specific member
err := suite.Assign(memberUserID, "reason")
err := suite.Remove(memberUserID, "reason")

// fluent updates
suite.SetName("new-role-name", "reason")
suite.SetColor(0x9B59B6)
suite.SetHoist(true)
suite.SetMentionable(false)
suite.SetPermissions(types.PermissionManageChannels | types.PermissionManageMessages)

// reorder role position in hierarchy
reorderedRoles, err := suite.SetPosition(5)

// delete role
err := suite.Delete("reason")
```

---

## Typed Errors

| Error Sentinel | Condition |
| :--- | :--- |
| `gdk.ErrInvalidRoleID` | Role ID is zero or unresolved |
| `gdk.ErrRoleNotFound` | Role was not found in guild role list |
| `gdk.ErrRoleOutranksBot` | Target role position is greater than or equal to bot's highest role |
| `gdk.ErrRoleOutranksUser` | Target role position is greater than or equal to invoker's highest role |
| `gdk.ErrCannotModifyEveryoneRole` | Attempted mutation or deletion of guild `@everyone` role |
| `gdk.ErrInvalidRoleAction` | Role action type is unrecognized |
| `gdk.ErrNotInGuild` | Attempted role actions in direct messages or outside a guild |
| `gdk.ErrNoRESTClient` | REST client is uninitialized on gateway context |

---

## Connected Topics

| Context | Destination | Description |
| :--- | :--- | :--- |
| **Modules Hub** | [Entity Modules Overview](index.md) | Universal options matrix and engine architecture |
| **Context** | [Context Guide](../context.md) | Interaction lifecycle, target resolution and role hierarchy |
| **Member** | [member.md](member.md) | Member moderation, profile roles and voice permissions |
| **Channel** | [channel.md](channel.md) | Channel creation, lock/unlock and permission overrides |
| **Respond** | [respond.md](respond.md) | Direct messages, role change notifications and replies |
| **Message** | [message.md](message.md) | Message pinning, reactions and older message edits |
| **API Reference** | [RoleSuite](../api-reference.md#fluent-action-suites) | Fluent companion suite and error sentinels |

