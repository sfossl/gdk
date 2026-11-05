[Docs Overview](../README.md#documentation) / Utilities

# Utilities

The `github.com/sfossl/gdk/util` package provides human duration parsing, Discord markdown timestamp formatting helpers, pointer utilities and decoupled file logging, all re-exported directly under `gdk.*` or `util.*`.

## Pointer Helper (`util.Ptr`)

`util.Ptr` takes any value and returns its pointer. This avoids temporary variables when setting optional bounds on options or commands:

```go
import "github.com/sfossl/gdk/util"

opt := gdk.Option{
	Type:        gdk.OptionTypeString,
	Name:        "reason",
	Description: "ban reason",
	MinLength:   util.Ptr(1),
	MaxLength:   util.Ptr(200),
}
```

## Logging Utility (`gdk.Logger`)

`gdk` ships with an optional, high-performance file logger designed for zero-config bots or production services needing clean terminal output and dedicated log streams.

```go
logger, err := gdk.NewLogger(gdk.LoggerConfig{
	Dir:     ".logs",           // default: ".logs"
	Format:  gdk.LogFormatText, // LogFormatText (.log), LogFormatJSON (.json), LogFormatCSV (.csv)
	Console: false,             // false keeps console silent; true mirrors to stderr
})
if err != nil {
	panic(err)
}
defer logger.Close()

// write to streams
logger.General("bot started successfully")
logger.Audit("user %d changed moderation role", userID)

// route panic recovery cleanly without console spam
app.Use(gdk.Guard(gdk.WithGuardLogger(logger)))
```

### Supported Log Formats

| Format | Extension | Output Structure |
| :--- | :--- | :--- |
| `gdk.LogFormatText` | `.log` | `YYYY-MM-DD HH:MM:SS [LEVEL] [category] message` |
| `gdk.LogFormatJSON` | `.json` | Single-line JSON objects with `time`, `level`, `category`, `message`, `stack` |
| `gdk.LogFormatCSV` | `.csv` | Standard RFC 4180 comma-separated records |

## Human Duration Parsing

`gdk.ParseDuration` parses compound human time strings into standard Go `time.Duration` values:

```go
// supports s, m, h, d, w units
duration, err := gdk.ParseDuration("1h 30m")
if err != nil {
	return err
}

// 2 days
banDuration, err := gdk.ParseDuration("2d")
```

Supported units:
- `s`: seconds
- `m`: minutes
- `h`: hours
- `d`: days (24 hours)
- `w`: weeks (7 days)

## Formatting Durations

Convert durations into human-readable strings:

```go
str := gdk.FormatDurationHuman(90 * time.Minute)
// output: "1 hour, 30 minutes"
```

## Discord Markdown Timestamps

Discord provides client-side dynamic timestamps using `<t:UNIX:STYLE>` markdown syntax. `gdk.Time` generates these tags cleanly via `.Discord(style)`:

```go
targetTime := time.Now().Add(10 * time.Minute)

// relative countdown: <t:1700000000:R>
rel := gdk.Time(targetTime).Discord("R")

// long date and time: <t:1700000000:F>
full := gdk.Time(targetTime).Discord("F")

// short time: <t:1700000000:t>
shortTime := gdk.Time(targetTime).Discord("t")

_, err := ctx.Respond(fmt.Sprintf("expires %s", rel))
return err
```

### Discord Timestamp Styles

| Style Flag | Alternate Name | Rendered Example |
| :--- | :--- | :--- |
| `"R"` | `"relative"` | in 10 minutes |
| `"t"` | `"time"` | 9:41 PM |
| `"T"` | `"long_time"` | 9:41:30 PM |
| `"d"` | `"date"` | 10/17/2026 |
| `"D"` | `"long_date"` | October 17, 2026 |
| `"f"` | `"short_date_time"` | October 17, 2026 9:41 PM |
| `"F"` | `"full"` | Saturday, October 17, 2026 9:41 PM |

---

## Connected Topics

| Context | Destination | Description |
| :--- | :--- | :--- |
| **Previous** | [Transports](transports.md) | Gateway WebSocket and HTTP Webhook execution models |
| **Next** | [API Reference](api-reference.md) | Exported packages, types, constructors and sentinels |
| **Commands** | [Commands](commands.md) | Using `util.Ptr` for optional bounds and schema limits |
| **Context** | [Context Guide](context.md) | Dynamic Discord timestamp replies and log outputs |
| **Modules** | [Member](modules/member.md), [Channel](modules/channel.md) | Duration strings for bans, timeouts and channel slowmode |
| **API Reference** | [Utilities API](api-reference.md#utilities) | `Time`, `Duration`, `Ptr` and `Logger` API contracts |
