package context

import (
	"context"
	"errors"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

var (
	ErrNotInGuild          = errors.New("interaction not in a guild")
	ErrNoRESTClient        = errors.New("rest client not available")
	ErrCannotModerateSelf  = errors.New("cannot moderate self")
	ErrCannotModerateOwner = errors.New("cannot moderate guild owner")
	ErrTargetOutranksUser  = errors.New("target outranks or equals user in role hierarchy")
	ErrTargetOutranksBot   = errors.New("target outranks or equals bot in role hierarchy")
	ErrInvalidTarget       = errors.New("invalid target member id")
)

func (c *Context) getGuildCached(ctx context.Context, guildID types.Snowflake) (*types.Guild, error) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	if c.cachedGuild != nil {
		return c.cachedGuild, nil
	}
	guild, err := c.REST.GetGuild(ctx, guildID)
	if err != nil {
		return nil, err
	}
	c.cachedGuild = guild
	if len(guild.Roles) > 0 {
		c.cachedRoles = guild.Roles
	}
	return guild, nil
}

func (c *Context) getRolesCached(ctx context.Context, guildID types.Snowflake) ([]types.Role, error) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	if len(c.cachedRoles) > 0 {
		return c.cachedRoles, nil
	}
	roles, err := c.REST.GetGuildRoles(ctx, guildID)
	if err != nil {
		return nil, err
	}
	c.cachedRoles = roles
	return roles, nil
}

func (c *Context) getBotMemberCached(ctx context.Context, guildID types.Snowflake) (*types.Member, error) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	if c.cachedBotMem != nil {
		return c.cachedBotMem, nil
	}
	bm, err := c.REST.GetGuildMember(ctx, guildID, c.AppID)
	if err != nil {
		return nil, err
	}
	c.cachedBotMem = bm
	return bm, nil
}

func (c *Context) CanModerate(targetMemberID types.Snowflake) (bool, error) {
	if targetMemberID == 0 {
		return false, ErrInvalidTarget
	}
	if c.GuildID == nil {
		return false, ErrNotInGuild
	}
	if c.REST == nil {
		return false, ErrNoRESTClient
	}

	guildID := *c.GuildID
	ctx := c.Context()

	guild, err := c.getGuildCached(ctx, guildID)
	if err != nil {
		return false, err
	}

	if targetMemberID == guild.OwnerID {
		return false, ErrCannotModerateOwner
	}

	invokerID := types.Snowflake(0)
	if c.Author != nil {
		invokerID = c.Author.ID
	}
	if invokerID != 0 && targetMemberID == invokerID {
		return false, ErrCannotModerateSelf
	}
	if c.AppID != 0 && targetMemberID == c.AppID {
		return false, ErrCannotModerateSelf
	}

	roles := guild.Roles
	if len(roles) == 0 {
		fetchedRoles, err := c.getRolesCached(ctx, guildID)
		if err != nil {
			return false, err
		}
		roles = fetchedRoles
	}

	rolePositions := make(map[types.Snowflake]int, len(roles))
	for _, r := range roles {
		rolePositions[r.ID] = r.Position
	}

	var targetRoles []types.Snowflake
	targetFound := false
	if c.Interaction != nil && c.Interaction.Data != nil && c.Interaction.Data.Resolved != nil && c.Interaction.Data.Resolved.Members != nil {
		if tm, ok := c.Interaction.Data.Resolved.Members[targetMemberID]; ok {
			targetRoles = tm.Roles
			targetFound = true
		}
	}
	if !targetFound {
		tm, err := c.REST.GetGuildMember(ctx, guildID, targetMemberID)
		if err != nil {
			if !rest.IsNotFound(err) {
				return false, err
			}
		} else {
			targetRoles = tm.Roles
		}
	}
	targetMaxPos := memberMaxRolePosition(targetRoles, rolePositions)

	if c.AppID == 0 || c.AppID != guild.OwnerID {
		botMember, err := c.getBotMemberCached(ctx, guildID)
		if err != nil {
			return false, err
		}
		botMaxPos := memberMaxRolePosition(botMember.Roles, rolePositions)
		if botMaxPos <= targetMaxPos {
			return false, ErrTargetOutranksBot
		}
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

	if invokerMaxPos <= targetMaxPos {
		return false, ErrTargetOutranksUser
	}

	return true, nil
}

func memberMaxRolePosition(roleIDs []types.Snowflake, positions map[types.Snowflake]int) int {
	maxPos := 0
	for _, id := range roleIDs {
		if pos, ok := positions[id]; ok {
			if pos > maxPos {
				maxPos = pos
			}
		}
	}
	return maxPos
}
