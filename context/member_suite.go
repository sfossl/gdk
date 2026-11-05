package context

import (
	"strings"

	"github.com/sfossl/gda/types"
)

type MemberSuite struct {
	ctx      *Context
	targetID types.Snowflake
}

func (c *Context) MemberSuite(targetID ...types.Snowflake) MemberSuite {
	tid := c.TargetID()
	if len(targetID) > 0 && targetID[0] != 0 {
		tid = targetID[0]
	}
	return MemberSuite{
		ctx:      c,
		targetID: tid,
	}
}

func (m MemberSuite) ID() types.Snowflake {
	return m.targetID
}

func (m MemberSuite) Get() (*types.Member, error) {
	if m.targetID == 0 {
		return nil, ErrInvalidTarget
	}
	if m.ctx.GuildID == nil {
		return nil, ErrNotInGuild
	}
	if m.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	return m.ctx.REST.GetGuildMember(m.ctx.Context(), *m.ctx.GuildID, m.targetID)
}

func (m MemberSuite) Kick(reason ...string) error {
	return m.ctx.Member(MemberOptions{
		Type:   ModKick,
		Target: m.targetID,
		Reason: strings.Join(reason, " "),
	})
}

func (m MemberSuite) Ban(deleteHistory any, reason ...string) error {
	return m.ctx.Member(MemberOptions{
		Type:           ModBan,
		Target:         m.targetID,
		DeleteMessages: deleteHistory,
		Reason:         strings.Join(reason, " "),
	})
}

func (m MemberSuite) Unban(reason ...string) error {
	return m.ctx.Member(MemberOptions{
		Type:   ModUnban,
		Target: m.targetID,
		Reason: strings.Join(reason, " "),
	})
}

func (m MemberSuite) Timeout(duration any, reason ...string) error {
	return m.ctx.Member(MemberOptions{
		Type:     ModTimeout,
		Target:   m.targetID,
		Duration: duration,
		Reason:   strings.Join(reason, " "),
	})
}

func (m MemberSuite) RemoveTimeout(reason ...string) error {
	return m.ctx.Member(MemberOptions{
		Type:   ModRemoveTimeout,
		Target: m.targetID,
		Reason: strings.Join(reason, " "),
	})
}

func (m MemberSuite) AddRole(roleID types.Snowflake, reason ...string) error {
	return m.ctx.Member(MemberOptions{
		Type:     ModAddRole,
		Target:   m.targetID,
		AddRoles: []types.Snowflake{roleID},
		Reason:   strings.Join(reason, " "),
	})
}

func (m MemberSuite) RemoveRole(roleID types.Snowflake, reason ...string) error {
	return m.ctx.Member(MemberOptions{
		Type:        ModRemoveRole,
		Target:      m.targetID,
		RemoveRoles: []types.Snowflake{roleID},
		Reason:      strings.Join(reason, " "),
	})
}

func (m MemberSuite) SetNick(nick string, reason ...string) error {
	return m.ctx.Member(MemberOptions{
		Target:   m.targetID,
		Nickname: &nick,
		Reason:   strings.Join(reason, " "),
	})
}

func (m MemberSuite) ResetNick(reason ...string) error {
	empty := ""
	return m.ctx.Member(MemberOptions{
		Target:   m.targetID,
		Nickname: &empty,
		Reason:   strings.Join(reason, " "),
	})
}

func (m MemberSuite) Mute(reason ...string) error {
	t := true
	return m.ctx.Member(MemberOptions{
		Target:    m.targetID,
		VoiceMute: &t,
		Reason:    strings.Join(reason, " "),
	})
}

func (m MemberSuite) Unmute(reason ...string) error {
	f := false
	return m.ctx.Member(MemberOptions{
		Target:    m.targetID,
		VoiceMute: &f,
		Reason:    strings.Join(reason, " "),
	})
}

func (m MemberSuite) Deaf(reason ...string) error {
	t := true
	return m.ctx.Member(MemberOptions{
		Target:    m.targetID,
		VoiceDeaf: &t,
		Reason:    strings.Join(reason, " "),
	})
}

func (m MemberSuite) Undeaf(reason ...string) error {
	f := false
	return m.ctx.Member(MemberOptions{
		Target:    m.targetID,
		VoiceDeaf: &f,
		Reason:    strings.Join(reason, " "),
	})
}

func (m MemberSuite) Move(channelID types.Snowflake, reason ...string) error {
	return m.ctx.Member(MemberOptions{
		Target:       m.targetID,
		VoiceChannel: &channelID,
		Reason:       strings.Join(reason, " "),
	})
}

func (m MemberSuite) Disconnect(reason ...string) error {
	var nilChan types.Snowflake
	return m.ctx.Member(MemberOptions{
		Target:       m.targetID,
		VoiceChannel: &nilChan,
		Reason:       strings.Join(reason, " "),
	})
}
