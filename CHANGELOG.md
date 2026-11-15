# Changelog

## v0.2.0 (2026-11-14)

- Declarative layout component engine: Containers, Sections, Separators, Media Galleries, File Cards, Thumbnails and Text Displays.
- Automatic message flag detection (`MessageFlagIsComponentsV2`) with mutual exclusion enforcement against legacy top-level embeds and content.
- Pre-flight component tree boundary validation: 40 total components limit, 4,000 total text characters ceiling, action row child element verification and non-interactive child rejection.
- Concurrency and lifecycle hardening: single atomic execution state machine replacing split boolean flags, preventing races between auto-deferral and modals.
- Exclusive mutex starvation fix: decoupled cache mutex acquisition from blocking network REST roundtrips during guild, role and member lookups.
- Moderation role hierarchy enforcement: client-side `CanModerate` pre-flight checks across all moderation actions (`Kick`, `Ban`, `Timeout`, `AddRole`, `RemoveRole`, `Nick`, `Voice`).
- Wire serialization parity: `AllowedMentions` empty slice `"parse": []` serialization ensuring Discord suppresses all mentions on the wire.
- Discord custom emoji parsing for interactive buttons and select menus (`<:name:id>` / `<a:name:id>`).
- Exported `WebhookHandler` for serverless runtimes (AWS Lambda, Google Cloud Run) and standard Go HTTP routers (`chi`, `gin`, `echo`).
- Webhook auto-deferral lifecycle synchronization and client disconnect cancellation propagation.
- Go 1.24 stdlib modernizations: adoption of `slices.Sort`, `slices.SortFunc`, `cmp.Compare` and `slices.Clone`.
- Re-exported complete root facade: `ChannelType` and all 13 `ChannelType*` constants, `Mod*` moderation constants, layout boundary constants and error sentinels in `gdk`.
- Zero-allocation `joinReason` optimization across `MemberSuite`, `ChannelSuite`, `RoleSuite` and `MessageSuite`.

## v0.1.0 (2026-11-06)

- Initial stable release of gdk application command framework.
- Differential command synchronization with deterministic SHA-256 schema hashing.
- Dual execution transports: persistent Gateway WebSocket and serverless HTTP Webhooks with Ed25519 signature verification.
- Slash command routing, typed option extraction with default fallbacks, subcommands and autocomplete routing.
- Declarative component builders for embeds, buttons, select menus, action rows and modals.
- Middleware pipeline with panic recovery, token-bucket cooldowns and role/permission guards.
- Stateless, client-verified fluent action suites for members, channels, roles and messages.
- Compound human duration parser and Discord markdown timestamp utilities.
- Pure Go 1.24 standard library implementation built directly on gda driver with zero third-party dependencies.

