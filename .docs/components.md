[Docs Overview](../README.md#documentation) / Components & Modals

# Components

The `gdk/component` package provides constructors and handlers for Discord interactive widgets: embeds, buttons, select menus, action rows and text inputs.

## Embed Builder

Construct rich embeds with limit validation:

```go
embed, err := gdk.NewEmbed().
	SetTitle("title").
	SetDescription("description").
	SetColor(0xFFFFFF).
	SetURL("https://sfossl.org").
	AddField("name", "value", false).
	Build()
if err != nil {
	return err
}

_, err = ctx.Respond(gdk.Options{
	Embed: embed,
})
return err
```

## Buttons and Rows

Buttons reside inside action rows. An action row supports up to 5 buttons.

### Unified Constructor
Use `gdk.Button(...)` with positional arguments or `gdk.ButtonOptions`:

```go
// positional constructor
btn1 := gdk.Button("primary", "button_id_1", "label")
btn2 := gdk.Button("danger", "button_id_2", "label")
btn3 := gdk.Button("link", "https://example.com", "label")

// structured constructor
btnOpt := gdk.Button(gdk.ButtonOptions{
	Style:    "success",
	CustomID: "button_id",
	Label:    "label",
	Disabled: false,
})

// monetization premium sku button
btnPrem := gdk.Button(gdk.ButtonOptions{
	Style: "premium",
	SKUID: 123456789012345678,
})
```

Supported styles (string or integer constant):
- `"primary"` / `gdk.ButtonStylePrimary`
- `"secondary"` / `gdk.ButtonStyleSecondary`
- `"success"` / `gdk.ButtonStyleSuccess`
- `"danger"` / `gdk.ButtonStyleDanger`
- `"link"` / `gdk.ButtonStyleLink`
- `"premium"` / `"sku"` / `gdk.ButtonStylePremium` (requires `SKUID`)

### Grouping in Action Rows
```go
row := gdk.ActionRow(btn1, btn2, btn3)

_, err = ctx.Respond(gdk.Options{
	Content:    "choose an option",
	Components: []gdk.Component{row},
})
```

### Auto-Packing Buttons
`gdk.PackRows(buttons...)` chunks any number of buttons into valid 5-button action rows automatically:

```go
rows := gdk.PackRows(btn1, btn2, btn3, btn4, btn5, btn6)
```

## Component Disabling & Lifecycle

Interactive components can be disabled on demand or automatically timed out:

### Instant Disabling on Action Complete
In a component interaction handler, disable all buttons and select menus on the clicked message immediately:

```go
func onConfirm(ctx *gdk.Context) error {
	return ctx.DisableComponents()
}
```

### Automatic Component Timeout
Automatically greys out all buttons and select menus on the response after a duration:

```go
_, err := ctx.Respond(gdk.Options{
	Embed:            promptEmbed,
	Components:       []gdk.Component{row},
	ComponentTimeout: 30 * time.Second,
})
```

For message auto-deletion (`DeleteAfter`), see [Context Response Options](context.md#response-lifecycle-timeouts).

## Select Menus

Select menus span the full width of an action row and must occupy their row alone:

```go
// string select menu
menu := gdk.StringSelect("menu_choice", "select an item",
	gdk.SelectOption("one", "val_1", "first option"),
	gdk.SelectOption("two", "val_2", "second option"),
)

// user select menu with range
userMenu := gdk.UserSelect("user_choice", "choose users",
	gdk.WithRange(1, 3),
)

row := gdk.ActionRow(menu)
```

## Structural Layout Components

GDK supports Discord's layout component structures: containers, sections, text displays, separators, media galleries, and file cards.

GDK automatically detects layout components passed to `ctx.Respond` or `ctx.UpdateMessage` and injects the required message flag (`gdk.MessageFlagIsComponentsV2` / `32768`), validating tree limits (40 components max, 4,000 text characters max, no nested containers, proper accessory placement) before dispatching to Discord.

> [!NOTE]
> When layout components are used, legacy top-level `Content` and `Embeds` are disabled by Discord. Pass all text through `gdk.Text(...)` inside or alongside your layout components. Mixing legacy content or embeds returns `ErrContentNotAllowedWithLayouts` or `ErrEmbedsNotAllowedWithLayouts`.

### Containers and Cards
Containers group layout elements with optional accent colors and spoiler flags:

```go
card := gdk.Container(
	gdk.Section(
		"### Title Header",
		"Subtitle or descriptive text",
	).WithAccessory(gdk.Thumbnail("https://example.com/logo.png")).Build(),
	gdk.Divider(),
	gdk.Text("Body paragraph content inside the container."),
).WithAccentColor(0x5865F2).Build()

// or shorthand container card:
card = gdk.ContainerCard(0x5865F2,
	gdk.Text("Simple container card content"),
)
```

### Sections and Accessories
Sections contain up to 3 text displays with an optional accessory (a Button or Thumbnail):

```go
// pure-text section (accessory is optional)
secText := gdk.Section("Section Header", "Supporting text description").Build()

btn := gdk.Button(gdk.ButtonOptions{
	Style:    "primary",
	CustomID: "btn_click",
	Label:    "Action",
})

// fluent builder with accessory
sec := gdk.Section(
	"First text line",
	"Second text line",
).WithAccessory(btn).Build()

// shorthand constructor
sec = gdk.SectionWithAccessory(btn, "First text line", "Second text line")
```

### Component IDs
Assign optional integer IDs to layout elements to identify nodes within layout responses:

```go
text := gdk.Text("Header content", 1)
sec := gdk.Section("Title").WithID(2).Build()
card := gdk.Container(sec).WithID(3).Build()
sep := gdk.Separator(gdk.WithSeparatorID(4))
thumb := gdk.Thumbnail("https://example.com/logo.png", gdk.WithThumbnailID(5))
file := gdk.FileCard("https://example.com/doc.pdf", gdk.WithFileID(6))
```

### Separators and Spacing
Use `gdk.Divider()` for a default divider or `gdk.Separator(...)` with options:

```go
divider := gdk.Divider()
largeSep := gdk.Separator(
	gdk.WithSpacing(gdk.SeparatorSpacingLarge),
	gdk.WithDivider(true),
)
```

### Media Galleries and Files
Display media items in a responsive grid or file download cards:

```go
gallery := gdk.MediaGallery(
	gdk.MediaItem("https://example.com/image1.png", gdk.WithItemDescription("Image 1")),
	gdk.MediaItem("https://example.com/image2.png", gdk.WithItemDescription("Image 2")),
)

file := gdk.FileCard("https://example.com/report.pdf",
	gdk.WithFileName("report.pdf"),
	gdk.WithFileSize(1048576),
)
```

## Text Inputs

Text inputs are used inside modal dialogs:

```go
// short single-line input
shortInput := gdk.TextInputShort("input_id", "label",
	gdk.WithPlaceholder("placeholder"),
	gdk.WithRequired(true),
	gdk.WithMaxLength(32),
)

// multi-line paragraph input
paragraphInput := gdk.TextInputParagraph("input_id", "label",
	gdk.WithPlaceholder("placeholder"),
	gdk.WithRequired(false),
	gdk.WithMaxLength(500),
)
```

## Interaction Routing

Register component callbacks globally in command modules or on `gdk.App`:

### Global Self-Registration
```go
func init() {
	gdk.RegisterComponent("btn_submit", handleSubmit)
}

func handleSubmit(ctx *gdk.Context) error {
	return ctx.UpdateMessage(gdk.Options{
		Content: "submitted",
	})
}
```

### App Registration
```go
app.Component("btn_submit", func(ctx *gdk.Context) error {
	_, err := ctx.UpdateMessage(gdk.Options{
		Content: "submitted",
	})
	return err
})
```

### Modals
```go
row := gdk.ActionRow(gdk.TextInputShort("feedback_input", "your feedback"))

err := ctx.Modal(gdk.Modal{
	CustomID:   "modal_feedback",
	Title:      "feedback form",
	Components: []gdk.Component{row},
})

app.Modal("modal_feedback", func(ctx *gdk.Context) error {
	_, err := ctx.Respond(gdk.Options{
		Content:   "feedback received",
		Ephemeral: true,
	})
	return err
})
```

---

## Connected Topics

| Context | Destination | Description |
| :--- | :--- | :--- |
| **Previous** | [Entity Modules](modules/index.md) | Universal options matrix and the 5 unified execution engines |
| **Next** | [Middleware](middleware.md) | Token-bucket rate limits, permissions and panic guards |
| **Context** | [Context Guide](context.md) | Response lifecycles, ephemeral messages and auto-deferral |
| **Respond** | [Respond Module](modules/respond.md) | Dispatching embeds, components and action rows via `ctx.Respond` |
| **API Reference** | [Components API](api-reference.md#components) | Embed builder, buttons, select menus and modal dialogs |
