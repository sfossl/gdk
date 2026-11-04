package context

import (
	"errors"
	"strings"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

type RoleAction int

const (
	RoleCreate RoleAction = iota + 1
	RoleEdit
	RoleDelete
)

var (
	ErrInvalidRoleID            = errors.New("invalid role id")
	ErrRoleNotFound             = errors.New("role not found")
	ErrRoleOutranksBot          = errors.New("role outranks or equals bot in role hierarchy")
	ErrRoleOutranksUser         = errors.New("role outranks or equals user in role hierarchy")
	ErrCannotModifyEveryoneRole = errors.New("cannot delete or modify @everyone role")
	ErrInvalidRoleAction        = errors.New("invalid role action type")
)

type RoleOptions struct {
	Type        RoleAction
	RoleID      types.Snowflake
	Name        string
	Color       int
	Permissions *types.Permission
	Hoist       *bool
	Mentionable *bool
	Reason      string
}

func parseRoleAction(v any) RoleAction {
	switch val := v.(type) {
	case RoleAction:
		return val
	case string:
		switch strings.ToLower(strings.TrimSpace(val)) {
		case "create":
			return RoleCreate
		case "edit":
			return RoleEdit
		case "delete":
			return RoleDelete
		}
	}
	return 0
}

func (c *Context) Role(args ...any) (*types.Role, error) {
	if c.GuildID == nil {
		return nil, ErrNotInGuild
	}
	if c.REST == nil {
		return nil, ErrNoRESTClient
	}
	if len(args) == 0 {
		return nil, ErrInvalidRoleAction
	}

	opts := RoleOptions{}
	parseRest := false
	switch v := args[0].(type) {
	case RoleOptions:
		opts = v
	case *RoleOptions:
		if v != nil {
			opts = *v
		}
	case Options:
		roleID := v.Role
		if roleID == 0 {
			roleID = v.Target
		}
		opts = RoleOptions{
			Type:        parseRoleAction(v.Type),
			RoleID:      roleID,
			Name:        v.Name,
			Color:       v.Color,
			Permissions: v.Permissions,
			Hoist:       v.Hoist,
			Mentionable: v.Mentionable,
			Reason:      v.Reason,
		}
	case *Options:
		if v != nil {
			roleID := v.Role
			if roleID == 0 {
				roleID = v.Target
			}
			opts = RoleOptions{
				Type:        parseRoleAction(v.Type),
				RoleID:      roleID,
				Name:        v.Name,
				Color:       v.Color,
				Permissions: v.Permissions,
				Hoist:       v.Hoist,
				Mentionable: v.Mentionable,
				Reason:      v.Reason,
			}
		}
	case RoleAction:
		opts.Type = v
		parseRest = true
	case types.Snowflake:
		opts.RoleID = v
		parseRest = true
	case int:
		opts.RoleID = types.Snowflake(v)
		parseRest = true
	case int64:
		opts.RoleID = types.Snowflake(v)
		parseRest = true
	case string:
		action := parseRoleAction(v)
		if action != 0 {
			opts.Type = action
		} else {
			opts.Name = v
		}
		parseRest = true
	default:
		return nil, ErrInvalidRoleAction
	}

	if parseRest {
		for _, arg := range args[1:] {
			switch a := arg.(type) {
			case types.Snowflake:
				opts.RoleID = a
			case int:
				if opts.RoleID == 0 && opts.Type != RoleCreate {
					opts.RoleID = types.Snowflake(a)
				} else {
					opts.Color = a
				}
			case int64:
				opts.RoleID = types.Snowflake(a)
			case RoleAction:
				opts.Type = a
			case string:
				if opts.Type == 0 {
					act := parseRoleAction(a)
					if act != 0 {
						opts.Type = act
						continue
					}
				}
				if opts.Name == "" && opts.Type != RoleDelete {
					opts.Name = a
				} else {
					opts.Reason = a
				}
			}
		}
	}

	switch opts.Type {
	case RoleCreate:
		payload := rest.RolePayload{}
		if opts.Name != "" {
			payload.Name = &opts.Name
		}
		if opts.Color != 0 {
			payload.Color = &opts.Color
		}
		if opts.Permissions != nil {
			payload.Permissions = opts.Permissions
		}
		if opts.Hoist != nil {
			payload.Hoist = opts.Hoist
		}
		if opts.Mentionable != nil {
			payload.Mentionable = opts.Mentionable
		}
		return c.REST.CreateGuildRole(c.Context(), *c.GuildID, payload, opts.Reason)

	case RoleEdit:
		if opts.RoleID == 0 {
			return nil, ErrInvalidRoleID
		}
		if ok, err := c.CanManageRole(opts.RoleID); !ok {
			return nil, err
		}
		payload := rest.RolePayload{}
		if opts.Name != "" {
			payload.Name = &opts.Name
		}
		if opts.Color != 0 {
			payload.Color = &opts.Color
		}
		if opts.Permissions != nil {
			payload.Permissions = opts.Permissions
		}
		if opts.Hoist != nil {
			payload.Hoist = opts.Hoist
		}
		if opts.Mentionable != nil {
			payload.Mentionable = opts.Mentionable
		}
		return c.REST.ModifyGuildRole(c.Context(), *c.GuildID, opts.RoleID, payload, opts.Reason)

	case RoleDelete:
		if opts.RoleID == 0 {
			return nil, ErrInvalidRoleID
		}
		if ok, err := c.CanManageRole(opts.RoleID); !ok {
			return nil, err
		}
		err := c.REST.DeleteGuildRole(c.Context(), *c.GuildID, opts.RoleID, opts.Reason)
		return nil, err

	default:
		return nil, ErrInvalidRoleAction
	}
}

func (c *Context) ManageRole(args ...any) (*types.Role, error) {
	return c.Role(args...)
}

type RoleSuite struct {
	ctx    *Context
	roleID types.Snowflake
}

func (c *Context) RoleSuite(roleID ...types.Snowflake) RoleSuite {
	var id types.Snowflake
	if len(roleID) > 0 {
		id = roleID[0]
	}
	return RoleSuite{
		ctx:    c,
		roleID: id,
	}
}

func (a RoleSuite) ID() types.Snowflake {
	return a.roleID
}

func (c *Context) CanManageRole(roleID types.Snowflake) (bool, error) {
	if roleID == 0 {
		return false, ErrInvalidRoleID
	}
	if c.GuildID == nil {
		return false, ErrNotInGuild
	}

	guildID := *c.GuildID
	if roleID == guildID {
		return false, ErrCannotModifyEveryoneRole
	}
	if c.REST == nil {
		return false, ErrNoRESTClient
	}

	ctx := c.Context()
	guild, err := c.getGuildCached(ctx, guildID)
	if err != nil {
		return false, err
	}

	roles := guild.Roles
	if len(roles) == 0 {
		fetchedRoles, err := c.getRolesCached(ctx, guildID)
		if err != nil {
			return false, err
		}
		roles = fetchedRoles
	}

	var targetRole *types.Role
	rolePositions := make(map[types.Snowflake]int, len(roles))
	for i := range roles {
		r := &roles[i]
		rolePositions[r.ID] = r.Position
		if r.ID == roleID {
			targetRole = r
		}
	}

	if targetRole == nil {
		return false, ErrRoleNotFound
	}
	rolePos := targetRole.Position

	if c.AppID == 0 || c.AppID != guild.OwnerID {
		botMember, err := c.getBotMemberCached(ctx, guildID)
		if err != nil {
			return false, err
		}
		botMaxPos := memberMaxRolePosition(botMember.Roles, rolePositions)
		if botMaxPos <= rolePos {
			return false, ErrRoleOutranksBot
		}
	}

	invokerID := types.Snowflake(0)
	if c.Author != nil {
		invokerID = c.Author.ID
	}
	if invokerID != 0 && invokerID == guild.OwnerID {
		return true, nil
	}

	var invokerRoles []types.Snowflake
	if c.InvokerMember != nil {
		invokerRoles = c.InvokerMember.Roles
	} else if invokerID != 0 {
		im, err := c.REST.GetGuildMember(ctx, guildID, invokerID)
		if err != nil {
			return false, err
		}
		invokerRoles = im.Roles
	}
	invokerMaxPos := memberMaxRolePosition(invokerRoles, rolePositions)

	if invokerMaxPos <= rolePos {
		return false, ErrRoleOutranksUser
	}

	return true, nil
}

func (a RoleSuite) Get() (*types.Role, error) {
	if a.roleID == 0 {
		return nil, ErrInvalidRoleID
	}
	if a.ctx.GuildID == nil {
		return nil, ErrNotInGuild
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}

	roles, err := a.ctx.REST.GetGuildRoles(a.ctx.Context(), *a.ctx.GuildID)
	if err != nil {
		return nil, err
	}
	for i := range roles {
		if roles[i].ID == a.roleID {
			return &roles[i], nil
		}
	}
	return nil, ErrRoleNotFound
}

func (a RoleSuite) Create(payload rest.RolePayload, reason ...string) (*types.Role, error) {
	if a.ctx.GuildID == nil {
		return nil, ErrNotInGuild
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.CreateGuildRole(a.ctx.Context(), *a.ctx.GuildID, payload, r)
}

func (a RoleSuite) Edit(payload rest.RolePayload, reason ...string) (*types.Role, error) {
	if ok, err := a.ctx.CanManageRole(a.roleID); !ok {
		return nil, err
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.ModifyGuildRole(a.ctx.Context(), *a.ctx.GuildID, a.roleID, payload, r)
}

func (a RoleSuite) Delete(reason ...string) error {
	if ok, err := a.ctx.CanManageRole(a.roleID); !ok {
		return err
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.DeleteGuildRole(a.ctx.Context(), *a.ctx.GuildID, a.roleID, r)
}

func (a RoleSuite) SetName(name string, reason ...string) (*types.Role, error) {
	return a.Edit(rest.RolePayload{Name: &name}, reason...)
}

func (a RoleSuite) SetColor(color int, reason ...string) (*types.Role, error) {
	return a.Edit(rest.RolePayload{Color: &color}, reason...)
}

func (a RoleSuite) SetPermissions(perms types.Permission, reason ...string) (*types.Role, error) {
	return a.Edit(rest.RolePayload{Permissions: &perms}, reason...)
}

func (a RoleSuite) SetHoist(hoist bool, reason ...string) (*types.Role, error) {
	return a.Edit(rest.RolePayload{Hoist: &hoist}, reason...)
}

func (a RoleSuite) SetMentionable(mentionable bool, reason ...string) (*types.Role, error) {
	return a.Edit(rest.RolePayload{Mentionable: &mentionable}, reason...)
}

func (a RoleSuite) Assign(memberID types.Snowflake, reason ...string) error {
	if memberID == 0 {
		return ErrInvalidTarget
	}
	if ok, err := a.ctx.CanManageRole(a.roleID); !ok {
		return err
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.AddGuildMemberRole(a.ctx.Context(), *a.ctx.GuildID, memberID, a.roleID, r)
}

func (a RoleSuite) Remove(memberID types.Snowflake, reason ...string) error {
	if memberID == 0 {
		return ErrInvalidTarget
	}
	if ok, err := a.ctx.CanManageRole(a.roleID); !ok {
		return err
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.RemoveGuildMemberRole(a.ctx.Context(), *a.ctx.GuildID, memberID, a.roleID, r)
}
