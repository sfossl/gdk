package context

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/util/time"
)

type ChannelAction int

const (
	ChannelCreate ChannelAction = iota + 1
	ChannelEdit
	ChannelDelete
	ChannelPurge
	ChannelLock
)

var (
	ErrInvalidChannelID     = errors.New("invalid channel id")
	ErrChannelNameRequired  = errors.New("channel name is required")
	ErrSlowmodeTooLong      = errors.New("slowmode duration exceeds maximum 6 hours")
	ErrSlowmodeNegative     = errors.New("slowmode duration cannot be negative")
	ErrInvalidPurgeCount    = errors.New("purge count must be between 2 and 100")
	ErrInvalidChannelAction = errors.New("invalid channel action type")
)

const MaxSlowmodeSeconds = 21600

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

func parseChannelAction(v any) ChannelAction {
	switch val := v.(type) {
	case ChannelAction:
		return val
	case string:
		switch strings.ToLower(strings.TrimSpace(val)) {
		case "create":
			return ChannelCreate
		case "edit":
			return ChannelEdit
		case "delete":
			return ChannelDelete
		case "purge":
			return ChannelPurge
		case "lock", "unlock":
			return ChannelLock
		}
	}
	return 0
}

func (c *Context) ManageChannel(args ...any) error {
	if len(args) == 0 {
		return ErrInvalidChannelAction
	}
	opts := ChannelOptions{}
	switch v := args[0].(type) {
	case ChannelOptions:
		opts = v
	case *ChannelOptions:
		if v != nil {
			opts = *v
		}
	case Options:
		locked := true
		if s, ok := v.Type.(string); ok && strings.ToLower(strings.TrimSpace(s)) == "unlock" {
			locked = false
		}
		opts = ChannelOptions{
			Type:       parseChannelAction(v.Type),
			ChannelID:  v.ResolveChannel(c),
			Name:       v.Name,
			Topic:      v.Topic,
			Slowmode:   v.Slowmode,
			PurgeCount: v.Count,
			Locked:     locked,
			Reason:     v.Reason,
		}
	case *Options:
		if v != nil {
			locked := true
			if s, ok := v.Type.(string); ok && strings.ToLower(strings.TrimSpace(s)) == "unlock" {
				locked = false
			}
			opts = ChannelOptions{
				Type:       parseChannelAction(v.Type),
				ChannelID:  v.ResolveChannel(c),
				Name:       v.Name,
				Topic:      v.Topic,
				Slowmode:   v.Slowmode,
				PurgeCount: v.Count,
				Locked:     locked,
				Reason:     v.Reason,
			}
		}
	case ChannelAction:
		opts.Type = v
		if v == ChannelLock {
			opts.Locked = true
		}
		for _, arg := range args[1:] {
			switch a := arg.(type) {
			case types.Snowflake:
				opts.ChannelID = a
			case int:
				opts.PurgeCount = a
			case string:
				if (opts.Type == ChannelCreate || opts.Type == ChannelEdit) && opts.Name == "" {
					opts.Name = a
				} else {
					opts.Reason = a
				}
			}
		}
	case string:
		opts.Type = parseChannelAction(v)
		if strings.ToLower(strings.TrimSpace(v)) == "lock" {
			opts.Locked = true
		}
		for _, arg := range args[1:] {
			switch a := arg.(type) {
			case types.Snowflake:
				opts.ChannelID = a
			case int:
				opts.PurgeCount = a
			case string:
				if (opts.Type == ChannelCreate || opts.Type == ChannelEdit) && opts.Name == "" {
					opts.Name = a
				} else {
					opts.Reason = a
				}
			}
		}
	default:
		return ErrInvalidChannelAction
	}

	chID := opts.ChannelID
	if chID == 0 {
		chID = c.ChannelID
	}
	if opts.Type != ChannelCreate && chID == 0 {
		return ErrInvalidChannelID
	}
	if c.REST == nil {
		return ErrNoRESTClient
	}

	switch opts.Type {
	case ChannelCreate:
		if c.GuildID == nil {
			return ErrNotInGuild
		}
		if strings.TrimSpace(opts.Name) == "" {
			return ErrChannelNameRequired
		}
		payload := rest.CreateChannelPayload{
			Name:     opts.Name,
			Topic:    opts.Topic,
			ParentID: opts.ParentID,
		}
		_, err := c.REST.CreateGuildChannel(c.Context(), *c.GuildID, payload, opts.Reason)
		return err

	case ChannelEdit:
		if chID == 0 {
			return ErrInvalidChannelID
		}
		payload := rest.ModifyChannelPayload{}
		if opts.Name != "" {
			payload.Name = &opts.Name
		}
		if opts.Topic != "" {
			payload.Topic = &opts.Topic
		}
		if opts.ParentID != nil {
			payload.ParentID = opts.ParentID
		}
		if opts.Slowmode != nil {
			sec, err := parseSlowmodeSeconds(opts.Slowmode)
			if err != nil {
				return err
			}
			payload.RateLimitPerUser = &sec
		}
		_, err := c.REST.ModifyChannel(c.Context(), chID, payload, opts.Reason)
		return err

	case ChannelDelete:
		if chID == 0 {
			return ErrInvalidChannelID
		}
		return c.REST.DeleteChannel(c.Context(), chID, opts.Reason)

	case ChannelLock:
		if chID == 0 {
			return ErrInvalidChannelID
		}
		if c.GuildID == nil {
			return ErrNotInGuild
		}
		var headers http.Header
		if opts.Reason != "" {
			headers = http.Header{"X-Audit-Log-Reason": []string{opts.Reason}}
		}
		if opts.Locked {
			overwrite := types.PermissionOverwrite{
				ID:   *c.GuildID,
				Type: 0,
				Deny: types.PermissionSendMessages,
			}
			_, err := c.REST.Do(c.Context(), rest.Request{
				Endpoint: rest.EndpointEditChannelPermissions(chID, *c.GuildID),
				Body:     overwrite,
				Headers:  headers,
			})
			return err
		}
		_, err := c.REST.Do(c.Context(), rest.Request{
			Endpoint: rest.EndpointDeleteChannelPermission(chID, *c.GuildID),
			Headers:  headers,
		})
		return err

	case ChannelPurge:
		if chID == 0 {
			return ErrInvalidChannelID
		}
		if opts.PurgeCount < 2 || opts.PurgeCount > 100 {
			return ErrInvalidPurgeCount
		}
		endpoint := rest.Endpoint{
			Method: "GET",
			Path:   fmt.Sprintf("/channels/%s/messages?limit=%d", chID, opts.PurgeCount),
			Key:    fmt.Sprintf("GET:/channels/%s/messages", chID),
		}
		data, err := c.REST.Do(c.Context(), rest.Request{Endpoint: endpoint})
		if err != nil {
			return err
		}
		var msgs []types.Message
		if err := json.Unmarshal(data, &msgs); err != nil {
			return err
		}
		if len(msgs) == 0 {
			return nil
		}
		if len(msgs) == 1 {
			return c.REST.DeleteMessage(c.Context(), chID, msgs[0].ID, opts.Reason)
		}
		msgIDs := make([]types.Snowflake, len(msgs))
		for i, m := range msgs {
			msgIDs[i] = m.ID
		}
		bulkEndpoint := rest.Endpoint{
			Method: "POST",
			Path:   fmt.Sprintf("/channels/%s/messages/bulk-delete", chID),
			Key:    fmt.Sprintf("POST:/channels/%s/messages/bulk-delete", chID),
		}
		bulkPayload := struct {
			Messages []types.Snowflake `json:"messages"`
		}{
			Messages: msgIDs,
		}
		req := rest.Request{
			Endpoint: bulkEndpoint,
			Body:     bulkPayload,
		}
		if opts.Reason != "" {
			req.Headers = http.Header{"X-Audit-Log-Reason": []string{opts.Reason}}
		}
		_, err = c.REST.Do(c.Context(), req)
		return err

	default:
		return ErrInvalidChannelAction
	}
}

type ChannelSuite struct {
	ctx       *Context
	channelID types.Snowflake
}

func (c *Context) Channel(channelID ...types.Snowflake) ChannelSuite {
	chID := c.ChannelID
	if len(channelID) > 0 && channelID[0] != 0 {
		chID = channelID[0]
	}
	return ChannelSuite{
		ctx:       c,
		channelID: chID,
	}
}

func (a ChannelSuite) ID() types.Snowflake {
	return a.channelID
}

func (a ChannelSuite) Get() (*types.Channel, error) {
	if a.channelID == 0 {
		return nil, ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	return a.ctx.REST.GetChannel(a.ctx.Context(), a.channelID)
}

func (a ChannelSuite) Delete(reason ...string) error {
	if a.channelID == 0 {
		return ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return ErrNoRESTClient
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.DeleteChannel(a.ctx.Context(), a.channelID, r)
}

func (a ChannelSuite) SetName(name string, reason ...string) (*types.Channel, error) {
	if a.channelID == 0 {
		return nil, ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	if strings.TrimSpace(name) == "" {
		return nil, ErrChannelNameRequired
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.ModifyChannel(a.ctx.Context(), a.channelID, rest.ModifyChannelPayload{
		Name: &name,
	}, r)
}

func (a ChannelSuite) SetTopic(topic string, reason ...string) (*types.Channel, error) {
	if a.channelID == 0 {
		return nil, ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.ModifyChannel(a.ctx.Context(), a.channelID, rest.ModifyChannelPayload{
		Topic: &topic,
	}, r)
}

func (a ChannelSuite) SetSlowmode(duration any, reason ...string) (*types.Channel, error) {
	if a.channelID == 0 {
		return nil, ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	sec, err := parseSlowmodeSeconds(duration)
	if err != nil {
		return nil, err
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.ModifyChannel(a.ctx.Context(), a.channelID, rest.ModifyChannelPayload{
		RateLimitPerUser: &sec,
	}, r)
}

func (a ChannelSuite) Edit(payload rest.ModifyChannelPayload, reason ...string) (*types.Channel, error) {
	if a.channelID == 0 {
		return nil, ErrInvalidChannelID
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.ModifyChannel(a.ctx.Context(), a.channelID, payload, r)
}

func (a ChannelSuite) Create(payload rest.CreateChannelPayload, reason ...string) (*types.Channel, error) {
	if a.ctx.GuildID == nil {
		return nil, ErrNotInGuild
	}
	if a.ctx.REST == nil {
		return nil, ErrNoRESTClient
	}
	if strings.TrimSpace(payload.Name) == "" {
		return nil, ErrChannelNameRequired
	}
	r := strings.Join(reason, " ")
	return a.ctx.REST.CreateGuildChannel(a.ctx.Context(), *a.ctx.GuildID, payload, r)
}

func parseSlowmodeSeconds(v any) (int, error) {
	if v == nil {
		return 0, nil
	}
	switch val := v.(type) {
	case int:
		if val < 0 {
			return 0, ErrSlowmodeNegative
		}
		if val > MaxSlowmodeSeconds {
			return 0, ErrSlowmodeTooLong
		}
		return val, nil
	case int64:
		if val < 0 {
			return 0, ErrSlowmodeNegative
		}
		if val > int64(MaxSlowmodeSeconds) {
			return 0, ErrSlowmodeTooLong
		}
		return int(val), nil
	case time.Duration:
		sec := int(val.Seconds())
		if sec < 0 {
			return 0, ErrSlowmodeNegative
		}
		if sec > MaxSlowmodeSeconds {
			return 0, ErrSlowmodeTooLong
		}
		return sec, nil
	case string:
		if val == "" || val == "0" || val == "0s" {
			return 0, nil
		}
		d, err := timeutil.ParseDuration(val)
		if err != nil {
			return 0, err
		}
		sec := int(d.Seconds())
		if sec < 0 {
			return 0, ErrSlowmodeNegative
		}
		if sec > MaxSlowmodeSeconds {
			return 0, ErrSlowmodeTooLong
		}
		return sec, nil
	default:
		return 0, ErrUnsupportedType
	}
}
