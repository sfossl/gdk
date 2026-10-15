package app

import (
	"errors"
	"time"

	"github.com/sfossl/gda/types"
)

var (
	ErrTokenRequired = errors.New("bot token is required")
	ErrAppIDRequired = errors.New("application id is required")
)

type Mode int

const (
	ModeGateway Mode = iota
	ModeWebhook
)

func (m Mode) String() string {
	switch m {
	case ModeGateway:
		return "gateway"
	case ModeWebhook:
		return "webhook"
	default:
		return "unknown"
	}
}

type Config struct {
	Token          string
	AppID          types.Snowflake
	Mode           Mode
	Intents        types.Intent
	DevMode        bool
	DevGuildID     types.Snowflake
	AutoDefer      bool
	AutoDeferDelay time.Duration
	AutoSync       bool
	Timeout        time.Duration
	WebhookAddr    string
	WebhookPath    string
	PublicKey      string
}

func DefaultConfig(token string, appID types.Snowflake) Config {
	return Config{
		Token:          token,
		AppID:          appID,
		Mode:           ModeGateway,
		Intents:        types.IntentsDefault,
		DevMode:        false,
		AutoDefer:      true,
		AutoDeferDelay: 2500 * time.Millisecond,
		AutoSync:       false,
		Timeout:        14*time.Minute + 50*time.Second,
		WebhookAddr:    ":8080",
		WebhookPath:    "/interactions",
	}
}

func (c Config) Validate() error {
	if c.Token == "" {
		return ErrTokenRequired
	}
	if c.AppID == 0 {
		return ErrAppIDRequired
	}
	return nil
}

type Option func(*Config)

func WithMode(mode Mode) Option {
	return func(c *Config) {
		c.Mode = mode
	}
}

func WithIntents(intents types.Intent) Option {
	return func(c *Config) {
		c.Intents = intents
	}
}

func WithDevMode(devMode bool, devGuildID ...types.Snowflake) Option {
	return func(c *Config) {
		c.DevMode = devMode
		if len(devGuildID) > 0 {
			c.DevGuildID = devGuildID[0]
		}
	}
}

func WithAutoDefer(autoDefer bool, delay ...time.Duration) Option {
	return func(c *Config) {
		c.AutoDefer = autoDefer
		if len(delay) > 0 && delay[0] > 0 {
			c.AutoDeferDelay = delay[0]
		}
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		if timeout > 0 {
			c.Timeout = timeout
		}
	}
}

func WithWebhook(addr, path, publicKey string) Option {
	return func(c *Config) {
		c.Mode = ModeWebhook
		if addr != "" {
			c.WebhookAddr = addr
		}
		if path != "" {
			c.WebhookPath = path
		}
		c.PublicKey = publicKey
	}
}

func WithAutoSync(autoSync bool) Option {
	return func(c *Config) {
		c.AutoSync = autoSync
	}
}
