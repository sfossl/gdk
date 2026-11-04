package app

import (
	"bytes"
	stdctx "context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/command"
	"github.com/sfossl/gdk/context"
)

func TestApp_StringRedactsToken(t *testing.T) {
	var nilApp *App
	if nilApp.String() != "<nil>" {
		t.Fatalf("expected <nil>, got %s", nilApp.String())
	}

	app, err := New("super-secret-bot-token-12345", 123)
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}
	s := app.String()
	if strings.Contains(s, "super-secret-bot-token-12345") {
		t.Fatalf("token leaked in App.String(): %s", s)
	}
	if !strings.Contains(s, "[REDACTED]") {
		t.Fatalf("expected [REDACTED] in App.String(): %s", s)
	}
}

func TestAutoDefer_SkipsAutocompleteAndPing(t *testing.T) {
	app, err := New("token", 123, WithAutoDefer(true, 10*time.Millisecond))
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	inAuto := &types.Interaction{
		ID:   1,
		Type: types.InteractionTypeApplicationCommandAutocomplete,
		Data: &types.InteractionData{
			Name: "testcmd",
			Options: []types.ApplicationCommandOption{
				{Name: "opt", Type: 3, Value: "foo", Focused: true},
			},
		},
	}
	ctxAuto := context.New(stdctx.Background(), inAuto, nil)

	var autocompleteExecuted atomic.Bool
	err = app.RegisterCommand(command.Command{
		Type:        command.CommandTypeChatInput,
		Name:        "testcmd",
		Description: "desc",
		AutocompleteHandler: func(c *context.Context) error {
			time.Sleep(30 * time.Millisecond)
			if c.IsDeferred() {
				t.Errorf("autocomplete should never be auto-deferred")
			}
			autocompleteExecuted.Store(true)
			return c.Autocomplete([]types.Choice{{Name: "foo", Value: "bar"}})
		},
	}, func(c *context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("failed to register command: %v", err)
	}

	err = app.Handle(ctxAuto)
	if err != nil {
		t.Fatalf("handle failed: %v", err)
	}
	if !autocompleteExecuted.Load() {
		t.Fatalf("autocomplete handler was not executed")
	}
	if ctxAuto.IsDeferred() {
		t.Fatalf("context was deferred for autocomplete interaction")
	}

	inPing := &types.Interaction{
		ID:   2,
		Type: types.InteractionTypePing,
	}
	ctxPing := context.New(stdctx.Background(), inPing, nil)
	err = app.Handle(ctxPing)
	if err != nil {
		t.Fatalf("ping handle failed: %v", err)
	}
	if ctxPing.IsDeferred() {
		t.Fatalf("context was deferred for ping interaction")
	}
}

func TestWebhookServer_ContextLifespan(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate ed25519 key: %v", err)
	}
	pubHex := hex.EncodeToString(pub)

	app, err := New("test-token", 12345,
		WithMode(ModeWebhook),
		WithWebhook(":0", "/interactions", pubHex),
		WithTimeout(2*time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	var bgContextAlive atomic.Bool
	var wg sync.WaitGroup
	wg.Add(1)

	err = app.RegisterCommand(command.Command{
		Type:        command.CommandTypeChatInput,
		Name:        "longtask",
		Description: "runs async work",
	}, func(c *context.Context) error {
		bgCtx := c.Context()
		_, _ = c.Respond("instant response")
		go func() {
			defer wg.Done()
			select {
			case <-time.After(100 * time.Millisecond):
				if bgCtx.Err() == nil {
					bgContextAlive.Store(true)
				}
			case <-bgCtx.Done():
				bgContextAlive.Store(false)
			}
		}()
		return nil
	})
	if err != nil {
		t.Fatalf("failed to register command: %v", err)
	}

	in := types.Interaction{
		ID:    types.Snowflake(100),
		AppID: types.Snowflake(12345),
		Type:  types.InteractionTypeApplicationCommand,
		Data: &types.InteractionData{
			Type: 1,
			Name: "longtask",
		},
	}
	body, _ := json.Marshal(in)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	var msg bytes.Buffer
	msg.WriteString(timestamp)
	msg.Write(body)
	sig := ed25519.Sign(priv, msg.Bytes())
	sigHex := hex.EncodeToString(sig)

	req := httptest.NewRequest(http.MethodPost, "/interactions", bytes.NewReader(body))
	req.Header.Set("X-Signature-Ed25519", sigHex)
	req.Header.Set("X-Signature-Timestamp", timestamp)
	rec := httptest.NewRecorder()

	handler, err := app.webhookHandler()
	if err != nil {
		t.Fatalf("failed to create webhook handler: %v", err)
	}

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	wg.Wait()

	if !bgContextAlive.Load() {
		t.Fatalf("background context was cancelled prematurely on handler exit")
	}
}

func TestWebhookServer_PanicCancelsContext(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate ed25519 key: %v", err)
	}
	pubHex := hex.EncodeToString(pub)

	app, err := New("test-token", 12345,
		WithMode(ModeWebhook),
		WithWebhook(":0", "/interactions", pubHex),
		WithTimeout(2*time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	var bgContextCancelled atomic.Bool
	var wg sync.WaitGroup
	wg.Add(1)

	err = app.RegisterCommand(command.Command{
		Type:        command.CommandTypeChatInput,
		Name:        "panictask",
		Description: "panics",
	}, func(c *context.Context) error {
		bgCtx := c.Context()
		go func() {
			defer wg.Done()
			select {
			case <-bgCtx.Done():
				bgContextCancelled.Store(true)
			case <-time.After(500 * time.Millisecond):
				bgContextCancelled.Store(false)
			}
		}()
		panic("handler crash")
	})
	if err != nil {
		t.Fatalf("failed to register command: %v", err)
	}

	in := types.Interaction{
		ID:    types.Snowflake(100),
		AppID: types.Snowflake(12345),
		Type:  types.InteractionTypeApplicationCommand,
		Data: &types.InteractionData{
			Type: 1,
			Name: "panictask",
		},
	}
	body, _ := json.Marshal(in)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	var msg bytes.Buffer
	msg.WriteString(timestamp)
	msg.Write(body)
	sig := ed25519.Sign(priv, msg.Bytes())
	sigHex := hex.EncodeToString(sig)

	req := httptest.NewRequest(http.MethodPost, "/interactions", bytes.NewReader(body))
	req.Header.Set("X-Signature-Ed25519", sigHex)
	req.Header.Set("X-Signature-Timestamp", timestamp)
	rec := httptest.NewRecorder()

	handler, err := app.webhookHandler()
	if err != nil {
		t.Fatalf("failed to create webhook handler: %v", err)
	}

	handler.ServeHTTP(rec, req)

	wg.Wait()

	if !bgContextCancelled.Load() {
		t.Fatalf("expected context to be cancelled after panic")
	}
}

func TestApp_SubcommandDirectHandlerRouting(t *testing.T) {
	app, err := New("test-token", 12345)
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	var getExecuted atomic.Bool
	var setExecuted atomic.Bool

	cmd := command.Command{
		Type:        command.CommandTypeChatInput,
		Name:        "config",
		Description: "config command",
		Subcommands: []command.Subcommand{
			{
				Name:        "get",
				Description: "get setting",
				Execute: func(c *context.Context) error {
					getExecuted.Store(true)
					return nil
				},
			},
			{
				Name:        "set",
				Description: "set setting",
				Execute: func(c *context.Context) error {
					setExecuted.Store(true)
					return nil
				},
			},
		},
	}

	err = app.RegisterCommand(cmd, nil)
	if err != nil {
		t.Fatalf("failed to register command: %v", err)
	}

	inGet := &types.Interaction{
		ID:   1,
		Type: types.InteractionTypeApplicationCommand,
		Data: &types.InteractionData{
			Name: "config",
			Type: 1,
			Options: []types.ApplicationCommandOption{
				{Name: "get", Type: int(types.OptionTypeSubcommand)},
			},
		},
	}
	ctxGet := context.New(stdctx.Background(), inGet, nil)
	err = app.Handle(ctxGet)
	if err != nil {
		t.Fatalf("handle get failed: %v", err)
	}
	if !getExecuted.Load() {
		t.Fatalf("subcommand get handler was not executed")
	}

	inSet := &types.Interaction{
		ID:   2,
		Type: types.InteractionTypeApplicationCommand,
		Data: &types.InteractionData{
			Name: "config",
			Type: 1,
			Options: []types.ApplicationCommandOption{
				{Name: "set", Type: int(types.OptionTypeSubcommand)},
			},
		},
	}
	ctxSet := context.New(stdctx.Background(), inSet, nil)
	err = app.Handle(ctxSet)
	if err != nil {
		t.Fatalf("handle set failed: %v", err)
	}
	if !setExecuted.Load() {
		t.Fatalf("subcommand set handler was not executed")
	}
}

func TestApp_GatewayContextLifecycle(t *testing.T) {
	app, err := New("token", 123, WithTimeout(1*time.Second))
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	var bgAlive atomic.Bool
	var wg sync.WaitGroup
	wg.Add(1)
	started := make(chan struct{})

	err = app.RegisterCommand(command.Command{
		Type:        command.CommandTypeChatInput,
		Name:        "gwtask",
		Description: "gw async task",
	}, func(c *context.Context) error {
		bgCtx := c.Context()
		close(started)
		go func() {
			defer wg.Done()
			select {
			case <-time.After(50 * time.Millisecond):
				if bgCtx.Err() == nil {
					bgAlive.Store(true)
				}
			case <-bgCtx.Done():
				bgAlive.Store(false)
			}
		}()
		return nil
	})
	if err != nil {
		t.Fatalf("failed to register command: %v", err)
	}

	in := &types.Interaction{
		ID:   10,
		Type: types.InteractionTypeApplicationCommand,
		Data: &types.InteractionData{
			Name: "gwtask",
			Type: 1,
		},
	}
	rawCtx, cancelRaw := stdctx.WithCancel(stdctx.Background())

	inPayload, _ := json.Marshal(in)
	app.Dispatcher.Dispatch(rawCtx, types.EventInteractionCreate, inPayload)

	select {
	case <-started:
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for gateway command handler to start")
	}

	cancelRaw()
	wg.Wait()

	if !bgAlive.Load() {
		t.Fatalf("gateway context was cancelled when connection closed")
	}
}

func TestApp_CloseCancelsInteractionContexts(t *testing.T) {
	app, err := New("token", 123, WithTimeout(5*time.Second))
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	var ctxCancelled atomic.Bool
	started := make(chan struct{})
	done := make(chan struct{})

	err = app.RegisterCommand(command.Command{
		Type:        command.CommandTypeChatInput,
		Name:        "shutdowntask",
		Description: "task during shutdown",
	}, func(c *context.Context) error {
		close(started)
		go func() {
			<-c.Context().Done()
			ctxCancelled.Store(true)
			close(done)
		}()
		return nil
	})
	if err != nil {
		t.Fatalf("failed to register command: %v", err)
	}

	in := &types.Interaction{
		ID:   11,
		Type: types.InteractionTypeApplicationCommand,
		Data: &types.InteractionData{
			Name: "shutdowntask",
			Type: 1,
		},
	}
	inPayload, _ := json.Marshal(in)
	app.Dispatcher.Dispatch(stdctx.Background(), types.EventInteractionCreate, inPayload)

	select {
	case <-started:
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for command to start")
	}

	if err := app.Close(); err != nil {
		t.Fatalf("unexpected error closing app: %v", err)
	}

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for context cancellation on app.Close")
	}

	if !ctxCancelled.Load() {
		t.Fatalf("expected context to be cancelled after app.Close")
	}
}

func TestContext_ExplicitCancel(t *testing.T) {
	rawCtx, cancel := stdctx.WithCancel(stdctx.Background())
	c := context.New(rawCtx, &types.Interaction{}, nil, cancel)

	if c.Context().Err() != nil {
		t.Fatalf("expected context to be active")
	}

	c.Cancel()

	if c.Context().Err() == nil {
		t.Fatalf("expected context to be cancelled after c.Cancel()")
	}
}

