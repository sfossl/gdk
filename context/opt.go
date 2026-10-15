package context

import (
	"encoding/json"
	"strconv"

	"github.com/sfossl/gda/types"
)

type OptionVal struct {
	val    any
	res    *types.ResolvedData
	exists bool
}

func (v OptionVal) Exists() bool {
	return v.exists
}

func (v OptionVal) String(fallback ...string) string {
	if !v.exists || v.val == nil {
		if len(fallback) > 0 {
			return fallback[0]
		}
		return ""
	}
	switch s := v.val.(type) {
	case string:
		return s
	case json.Number:
		return s.String()
	default:
		return ""
	}
}

func (v OptionVal) Int(fallback ...int) int {
	if !v.exists || v.val == nil {
		if len(fallback) > 0 {
			return fallback[0]
		}
		return 0
	}
	switch n := v.val.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case json.Number:
		if i, err := n.Int64(); err == nil {
			return int(i)
		}
	case string:
		if i, err := strconv.Atoi(n); err == nil {
			return i
		}
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	return 0
}

func (v OptionVal) Int64(fallback ...int64) int64 {
	if !v.exists || v.val == nil {
		if len(fallback) > 0 {
			return fallback[0]
		}
		return 0
	}
	switch n := v.val.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	case json.Number:
		if i, err := n.Int64(); err == nil {
			return i
		}
	case string:
		if i, err := strconv.ParseInt(n, 10, 64); err == nil {
			return i
		}
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	return 0
}

func (v OptionVal) Float(fallback ...float64) float64 {
	if !v.exists || v.val == nil {
		if len(fallback) > 0 {
			return fallback[0]
		}
		return 0
	}
	switch n := v.val.(type) {
	case float64:
		return n
	case float32:
		return float64(n)
	case int:
		return float64(n)
	case int64:
		return float64(n)
	case json.Number:
		if f, err := n.Float64(); err == nil {
			return f
		}
	case string:
		if f, err := strconv.ParseFloat(n, 64); err == nil {
			return f
		}
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	return 0
}

func (v OptionVal) Bool(fallback ...bool) bool {
	if !v.exists || v.val == nil {
		if len(fallback) > 0 {
			return fallback[0]
		}
		return false
	}
	switch b := v.val.(type) {
	case bool:
		return b
	case string:
		if parsed, err := strconv.ParseBool(b); err == nil {
			return parsed
		}
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	return false
}

func (v OptionVal) Snowflake(fallback ...types.Snowflake) types.Snowflake {
	if !v.exists || v.val == nil {
		if len(fallback) > 0 {
			return fallback[0]
		}
		return 0
	}
	switch s := v.val.(type) {
	case string:
		if id, err := strconv.ParseUint(s, 10, 64); err == nil {
			return types.Snowflake(id)
		}
	case float64:
		return types.Snowflake(uint64(s))
	case int64:
		return types.Snowflake(uint64(s))
	case int:
		return types.Snowflake(uint64(s))
	case uint64:
		return types.Snowflake(s)
	case json.Number:
		if id, err := strconv.ParseUint(s.String(), 10, 64); err == nil {
			return types.Snowflake(id)
		}
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	return 0
}

func (v OptionVal) User() *types.User {
	if v.res == nil || v.res.Users == nil {
		return nil
	}
	sf := v.Snowflake()
	if sf == 0 {
		return nil
	}
	if u, ok := v.res.Users[sf]; ok {
		return &u
	}
	return nil
}

func (v OptionVal) Member() *types.Member {
	if v.res == nil || v.res.Members == nil {
		return nil
	}
	sf := v.Snowflake()
	if sf == 0 {
		return nil
	}
	if m, ok := v.res.Members[sf]; ok {
		if m.User == nil && v.res.Users != nil {
			if u, okU := v.res.Users[sf]; okU {
				m.User = &u
			}
		}
		return &m
	}
	return nil
}

func (v OptionVal) Role() *types.Role {
	if v.res == nil || v.res.Roles == nil {
		return nil
	}
	sf := v.Snowflake()
	if sf == 0 {
		return nil
	}
	if r, ok := v.res.Roles[sf]; ok {
		return &r
	}
	return nil
}

func (v OptionVal) Channel() *types.Channel {
	if v.res == nil || v.res.Channels == nil {
		return nil
	}
	sf := v.Snowflake()
	if sf == 0 {
		return nil
	}
	if ch, ok := v.res.Channels[sf]; ok {
		return &ch
	}
	return nil
}

func (v OptionVal) Attachment() *types.Attachment {
	if v.res == nil || v.res.Attachments == nil {
		return nil
	}
	sf := v.Snowflake()
	if sf == 0 {
		return nil
	}
	if a, ok := v.res.Attachments[sf]; ok {
		return &a
	}
	return nil
}

func (c *Context) Opt(name string) OptionVal {
	c.optsMu.RLock()
	if c.optsMap != nil {
		val := c.optsMap[name]
		c.optsMu.RUnlock()
		return val
	}
	c.optsMu.RUnlock()

	c.optsMu.Lock()
	defer c.optsMu.Unlock()

	if c.optsMap != nil {
		return c.optsMap[name]
	}

	c.optsMap = make(map[string]OptionVal)
	if c.Interaction == nil || c.Interaction.Data == nil {
		return OptionVal{}
	}

	res := c.Interaction.Data.Resolved
	extractLeafOptions(c.Interaction.Data.Options, res, c.optsMap)

	return c.optsMap[name]
}

func extractLeafOptions(opts []types.ApplicationCommandOption, res *types.ResolvedData, target map[string]OptionVal) {
	for _, opt := range opts {
		switch opt.Type {
		case 1, 2:
			extractLeafOptions(opt.Options, res, target)
		default:
			target[opt.Name] = OptionVal{
				val:    opt.Value,
				res:    res,
				exists: true,
			}
		}
	}
}

