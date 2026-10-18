package middleware

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/context"
	"github.com/sfossl/gdk/util/time"
)

var (
	ErrOnCooldown       = errors.New("action is on cooldown")
	ErrInvalidCooldown  = errors.New("invalid cooldown configuration")
	ErrMissingGuildInfo = errors.New("interaction missing guild information")
)

type Scope int

const (
	ScopeUser Scope = iota
	ScopeChannel
	ScopeGuild
	ScopeGlobal
)

func (s Scope) String() string {
	switch s {
	case ScopeUser:
		return "user"
	case ScopeChannel:
		return "channel"
	case ScopeGuild:
		return "guild"
	case ScopeGlobal:
		return "global"
	default:
		return "unknown"
	}
}

type CooldownError struct {
	Scope      Scope
	Remaining  time.Duration
	RetryAfter time.Time
}

func (e *CooldownError) Error() string {
	return "action is on cooldown"
}

func (e *CooldownError) Is(target error) bool {
	return target == ErrOnCooldown
}

type Policy struct {
	Scope    Scope
	Duration time.Duration
	Burst    int
}

func UserCooldown(duration any, burst ...int) Policy {
	return newPolicy(ScopeUser, duration, burst...)
}

func ChannelCooldown(duration any, burst ...int) Policy {
	return newPolicy(ScopeChannel, duration, burst...)
}

func GuildCooldown(duration any, burst ...int) Policy {
	return newPolicy(ScopeGuild, duration, burst...)
}

func GlobalCooldown(duration any, burst ...int) Policy {
	return newPolicy(ScopeGlobal, duration, burst...)
}

func newPolicy(scope Scope, duration any, burst ...int) Policy {
	dur := parseCooldownDuration(duration)
	b := 1
	if len(burst) > 0 && burst[0] > 1 {
		b = burst[0]
	}
	return Policy{
		Scope:    scope,
		Duration: dur,
		Burst:    b,
	}
}

func parseCooldownDuration(v any) time.Duration {
	if v == nil {
		return time.Second
	}
	switch val := v.(type) {
	case time.Duration:
		if val <= 0 {
			return time.Second
		}
		return val
	case int:
		if val <= 0 {
			return time.Second
		}
		return time.Duration(val) * time.Second
	case int64:
		if val <= 0 {
			return time.Second
		}
		return time.Duration(val) * time.Second
	case string:
		item := timeutil.New(val)
		if item.Error() == nil && item.Duration() > 0 {
			return item.Duration()
		}
		return time.Second
	default:
		return time.Second
	}
}

type bucket struct {
	tokens     float64
	capacity   float64
	rate       float64
	duration   time.Duration
	lastRefill time.Time
}

type shard struct {
	mu        sync.Mutex
	buckets   map[string]*bucket
	lastSweep time.Time
}

const shardCount = 64

type CooldownStore struct {
	shards [shardCount]*shard
}

func NewCooldownStore() *CooldownStore {
	store := &CooldownStore{}
	for i := 0; i < shardCount; i++ {
		store.shards[i] = &shard{
			buckets: make(map[string]*bucket),
		}
	}
	return store
}

func (s *CooldownStore) getShard(key string) *shard {
	var h uint32 = 2166136261
	for i := 0; i < len(key); i++ {
		h ^= uint32(key[i])
		h *= 16777619
	}
	return s.shards[h%shardCount]
}

func (s *CooldownStore) Take(key string, policy Policy) (bool, time.Duration) {
	if s == nil {
		return true, 0
	}
	if policy.Duration <= 0 {
		policy.Duration = time.Second
	}
	if policy.Burst <= 0 {
		policy.Burst = 1
	}

	sh := s.getShard(key)
	sh.mu.Lock()
	defer sh.mu.Unlock()

	now := time.Now()

	if len(sh.buckets) > 32 || (!sh.lastSweep.IsZero() && now.Sub(sh.lastSweep) >= time.Minute) || (sh.lastSweep.IsZero() && len(sh.buckets) > 0) {
		sh.lastSweep = now
		for k, b := range sh.buckets {
			if now.Sub(b.lastRefill) >= b.duration {
				delete(sh.buckets, k)
			}
		}
	}

	b, ok := sh.buckets[key]
	if !ok {
		capVal := float64(policy.Burst)
		rateVal := capVal / policy.Duration.Seconds()
		sh.buckets[key] = &bucket{
			tokens:     capVal - 1.0,
			capacity:   capVal,
			rate:       rateVal,
			duration:   policy.Duration,
			lastRefill: now,
		}
		return true, 0
	}

	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.rate
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.lastRefill = now

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true, 0
	}

	needed := 1.0 - b.tokens
	waitSec := needed / b.rate
	waitTime := time.Duration(waitSec * float64(time.Second))
	if waitTime < time.Millisecond {
		waitTime = time.Millisecond
	}
	return false, waitTime
}

func (s *CooldownStore) Sweep() {
	if s == nil {
		return
	}
	now := time.Now()
	for _, sh := range s.shards {
		sh.mu.Lock()
		for k, b := range sh.buckets {
			if now.Sub(b.lastRefill) >= b.duration {
				delete(sh.buckets, k)
			}
		}
		sh.mu.Unlock()
	}
}

type CooldownConfig struct {
	Store       *CooldownStore
	Policies    []Policy
	AutoRespond bool
	MessageFn   func(remaining time.Duration, scope Scope) string
	KeyPrefix   string
}

type CooldownOption func(*CooldownConfig)

func WithCooldownStore(store *CooldownStore) CooldownOption {
	return func(c *CooldownConfig) {
		c.Store = store
	}
}

func WithCooldownAutoRespond(autoRespond bool) CooldownOption {
	return func(c *CooldownConfig) {
		c.AutoRespond = autoRespond
	}
}

func WithCooldownMessage(fn func(remaining time.Duration, scope Scope) string) CooldownOption {
	return func(c *CooldownConfig) {
		c.MessageFn = fn
	}
}

func WithKeyPrefix(prefix string) CooldownOption {
	return func(c *CooldownConfig) {
		c.KeyPrefix = prefix
	}
}

func Cooldown(policies []Policy, opts ...CooldownOption) Middleware {
	cfg := CooldownConfig{
		Store:       NewCooldownStore(),
		Policies:    policies,
		AutoRespond: true,
		MessageFn: func(remaining time.Duration, scope Scope) string {
			formatted := timeutil.FormatDuration(remaining)
			return fmt.Sprintf("You are on cooldown. Please wait %s before trying again.", formatted)
		},
		KeyPrefix: "cmd",
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}

	return func(next Handler) Handler {
		return func(ctx *context.Context) error {
			if ctx == nil {
				if next != nil {
					return next(ctx)
				}
				return nil
			}

			now := time.Now()
			for _, policy := range cfg.Policies {
				if policy.Duration < 0 || policy.Burst < 0 {
					return ErrInvalidCooldown
				}

				key, err := resolveCooldownKey(cfg.KeyPrefix, ctx, policy.Scope)
				if err != nil {
					if errors.Is(err, ErrInvalidCooldown) {
						return ErrInvalidCooldown
					}
					continue
				}

				ok, remaining := cfg.Store.Take(key, policy)
				if !ok {
					cooldownErr := &CooldownError{
						Scope:      policy.Scope,
						Remaining:  remaining,
						RetryAfter: now.Add(remaining),
					}

					if cfg.AutoRespond && (!ctx.IsAcknowledged() || ctx.IsDeferred()) {
						msg := cfg.MessageFn(remaining, policy.Scope)
						_, _ = ctx.Respond(context.Response{
							Content:   msg,
							Ephemeral: true,
						})
					}

					return cooldownErr
				}
			}

			if next != nil {
				return next(ctx)
			}
			return nil
		}
	}
}

func resolveCooldownKey(prefix string, ctx *context.Context, scope Scope) (string, error) {
	if ctx == nil {
		return "", errors.New("nil context")
	}
	switch scope {
	case ScopeUser:
		var uid types.Snowflake
		if ctx.Author != nil {
			uid = ctx.Author.ID
		} else if ctx.Member != nil && ctx.Member.User != nil {
			uid = ctx.Member.User.ID
		}
		if uid == 0 {
			return "", errors.New("missing user id")
		}
		return fmt.Sprintf("%s/u/%d", prefix, uid), nil

	case ScopeChannel:
		if ctx.ChannelID == 0 {
			return "", errors.New("missing channel id")
		}
		return fmt.Sprintf("%s/c/%d", prefix, ctx.ChannelID), nil

	case ScopeGuild:
		if ctx.GuildID != nil && *ctx.GuildID != 0 {
			return fmt.Sprintf("%s/g/%d", prefix, *ctx.GuildID), nil
		}
		if ctx.ChannelID != 0 {
			return fmt.Sprintf("%s/dm/%d", prefix, ctx.ChannelID), nil
		}
		return "", ErrMissingGuildInfo

	case ScopeGlobal:
		return fmt.Sprintf("%s/global", prefix), nil

	default:
		return "", ErrInvalidCooldown
	}
}
