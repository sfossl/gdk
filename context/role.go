package context

import (
	"errors"
	"strings"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

var (
	ErrInvalidRoleID            = errors.New("invalid role id")
	ErrRoleNotFound             = errors.New("role not found")
	ErrRoleOutranksBot          = errors.New("role outranks or equals bot in role hierarchy")
	ErrRoleOutranksUser         = errors.New("role outranks or equals user in role hierarchy")
	ErrCannotModifyEveryoneRole = errors.New("cannot delete or modify @everyone role")
)

type RoleAction struct {
	ctx    *Context
	roleID types.Snowflake
}

func (c *Context) Role(roleID ...types.Snowflake) RoleAction {
	var id types.Snowflake
	if len(roleID) > 0 {
		id = roleID[0]
	}
	return RoleAction{
		ctx:    c,
		roleID: id,
	}
}

func (c *Context) ManageRole(roleID ...types.Snowflake) RoleAction {
	return c.Role(roleID...)
}

func (a RoleAction) ID() types.Snowflake {
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
	guild, err := c.REST.GetGuild(ctx, guildID)
	if err != nil {
		return false, err
	}

	roles := guild.Roles
	if len(roles) == 0 {
		fetchedRoles, err := c.REST.GetGuildRoles(ctx, guildID)
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

	botMember, err := c.REST.GetGuildMember(ctx, guildID, c.AppID)
	if err != nil {
		return false, err
	}
	botMaxPos := memberMaxRolePosition(botMember.Roles, rolePositions)

	if c.AppID == 0 || c.AppID != guild.OwnerID {
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
	if c.Member != nil {
		invokerRoles = c.Member.Roles
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

func (a RoleAction) Get() (*types.Role, error) {
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

func (a RoleAction) Create(payload rest.RolePayload, reason ...string) (*types.Role, error) {
	if a.ctx.GuildID == nil {
		return nil, ErrNotInGuild
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.CreateGuildRole(a.ctx.Context(), *a.ctx.GuildID, payload, r)
}

func (a RoleAction) Edit(payload rest.RolePayload, reason ...string) (*types.Role, error) {
	if ok, err := a.ctx.CanManageRole(a.roleID); !ok {
		return nil, err
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.ModifyGuildRole(a.ctx.Context(), *a.ctx.GuildID, a.roleID, payload, r)
}

func (a RoleAction) Delete(reason ...string) error {
	if ok, err := a.ctx.CanManageRole(a.roleID); !ok {
		return err
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.DeleteGuildRole(a.ctx.Context(), *a.ctx.GuildID, a.roleID, r)
}

func (a RoleAction) SetName(name string, reason ...string) (*types.Role, error) {
	return a.Edit(rest.RolePayload{Name: &name}, reason...)
}

func (a RoleAction) SetColor(color int, reason ...string) (*types.Role, error) {
	return a.Edit(rest.RolePayload{Color: &color}, reason...)
}

func (a RoleAction) SetPermissions(perms types.Permission, reason ...string) (*types.Role, error) {
	return a.Edit(rest.RolePayload{Permissions: &perms}, reason...)
}

func (a RoleAction) SetHoist(hoist bool, reason ...string) (*types.Role, error) {
	return a.Edit(rest.RolePayload{Hoist: &hoist}, reason...)
}

func (a RoleAction) SetMentionable(mentionable bool, reason ...string) (*types.Role, error) {
	return a.Edit(rest.RolePayload{Mentionable: &mentionable}, reason...)
}

func (a RoleAction) Assign(memberID types.Snowflake, reason ...string) error {
	if memberID == 0 {
		return ErrInvalidTarget
	}
	if ok, err := a.ctx.CanManageRole(a.roleID); !ok {
		return err
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.AddGuildMemberRole(a.ctx.Context(), *a.ctx.GuildID, memberID, a.roleID, r)
}

func (a RoleAction) Remove(memberID types.Snowflake, reason ...string) error {
	if memberID == 0 {
		return ErrInvalidTarget
	}
	if ok, err := a.ctx.CanManageRole(a.roleID); !ok {
		return err
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.RemoveGuildMemberRole(a.ctx.Context(), *a.ctx.GuildID, memberID, a.roleID, r)
}
