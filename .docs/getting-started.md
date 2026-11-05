[Docs Overview](../README.md#documentation) / Getting Started

# Getting Started

`gdk` is an application command framework built on top of `gda`. It provides slash command routing, differential command synchronization, token-bucket rate limiting, UI component handling and dual execution transports (Gateway WebSocket or serverless HTTP Webhooks).

## Install

Requires Go 1.24 or higher:

```bash
go get github.com/sfossl/gdk
```

## Quickstart

A recommended project layout separates command definitions from application bootstrapping:

```
bot/
├── commands/
│   └── info.go
├── go.mod
└── main.go
```

### `commands/info.go`

Define command schemas and interaction handlers as decoupled units:

```go
package commands

import (
	"fmt"
	"time"

	"github.com/sfossl/gdk"
)

var Info = gdk.Command{
	Name:        "info",
	Description: "Display user account details and guild profile",
	Cooldown:    5 * time.Second,
	Options: []gdk.Option{
		gdk.UserOption("target", "User to inspect").WithRequired(),
	},
	Execute: HandleInfo,
}

func HandleInfo(ctx *gdk.Context) error {
	user := ctx.Opt("target").User()
	if user == nil {
		user = ctx.Author
	}

	embed := gdk.NewEmbed().
		SetTitle("User Information").
		SetDescription(fmt.Sprintf("**Username**: %s\n**ID**: %s", user.Username, user.ID)).
		SetColor(gdk.ColorBlurple).
		ToEmbed()

	_, err := ctx.Respond(gdk.Options{
		Embeds: []gdk.Embed{embed},
	})
	return err
}
```

### `main.go`

Initialize the application, register commands and run the execution loop:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bot/commands"
	"github.com/sfossl/gdk"
)

func main() {
	token := os.Getenv("TOKEN")
	appID, err := gdk.ParseSnowflake(os.Getenv("APP_ID"))
	if err != nil {
		log.Fatalf("invalid app id: %v", err)
	}

	devGuildID, _ := gdk.ParseSnowflake(os.Getenv("GUILD_ID"))

	app, err := gdk.NewApp(
		token,
		appID,
		gdk.WithDevMode(true, devGuildID),
		gdk.WithAutoSync(true),
		gdk.WithAutoDefer(true, 2500*time.Millisecond),
	)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	if err := app.Command(commands.Info); err != nil {
		log.Fatalf("failed to register command: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Println("starting bot...")
	if err := app.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("runtime error: %v", err)
	}

	_ = app.Close()
}
```

## Options

| Option | Signature | Default | Description |
| :--- | :--- | :--- | :--- |
| `WithDevMode` | `WithDevMode(bool, ...gdk.Snowflake)` | `false, 0` | Routes command synchronization to a development guild for instant updates. |
| `WithAutoSync` | `WithAutoSync(bool)` | `false` | Automatically runs differential command synchronization on startup. |
| `WithAutoDefer` | `WithAutoDefer(bool, ...time.Duration)` | `true, 2.5s` | Automatically defers interaction responses before Discord's 3-second deadline expires. |
| `WithMode` | `WithMode(Mode)` | `ModeGateway` | Selects `ModeGateway` (persistent WebSocket) or `ModeWebhook` (HTTP server). |
| `WithWebhook` | `WithWebhook(addr, path, publicKey)` | `":8080", "/interactions", ""` | Configures HTTP webhook endpoint and Ed25519 signature verification. |
| `WithIntents` | `WithIntents(gdk.Intent)` | `IntentsDefault` | Gateway event subscription bitmask. |
| `WithTimeout` | `WithTimeout(time.Duration)` | `14m50s` | Maximum interaction lifecycle duration before context cancellation. |

## Environments

Discord caches global slash commands across its edge network for up to an hour. To test commands instantly during development:

1. Provide `GUILD_ID` in your environment.
2. Pass `gdk.WithDevMode(true, devGuildID)` to `gdk.NewApp`.
3. Set `gdk.WithAutoSync(true)`.

When `DevMode` is enabled, all commands register directly to that guild and update in Discord within seconds.

## Lifecycle

### Starting the Application
`app.Run(ctx)` starts the execution engine, completes the initial handshake, syncs registered commands if `WithAutoSync(true)` is set and blocks until the context terminates:

```go
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
defer cancel()

if err := app.Run(ctx); err != nil && err != context.Canceled {
	log.Printf("runtime error: %v", err)
}
```

### Clean Shutdown
`app.Close()` shuts down active network sockets, closes idle HTTP connections and terminates background workers:

```go
defer app.Close()
```

---

## Connected Topics

| Context | Destination | Description |
| :--- | :--- | :--- |
| **Index** | [Docs Overview](../README.md#documentation) | Main documentation directory and architecture summary |
| **Next** | [Commands](commands.md) | Slash commands, subcommands and context menus |
| **Transports** | [Transports](transports.md) | Gateway WebSocket and HTTP Webhook execution models |
| **Sync** | [Synchronization](sync.md) | Differential command hashing and deployment scopes |
| **API Reference** | [Application API](api-reference.md#application) | Constructors, configuration options and lifecycle methods |
