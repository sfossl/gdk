package app

import (
	"bytes"
	stdctx "context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/command"
	"github.com/sfossl/gdk/context"
	"github.com/sfossl/gdk/middleware"
)

func TestConfig_DefaultsAndOptions(t *testing.T) {
	cfg := DefaultConfig("test-token", 123456789)
	if cfg.Token != "test-token" {
		t.Fatalf("expected token test-token, got %s", cfg.Token)
	}
	if cfg.AppID != 123456789 {
		t.Fatalf("expected app id 123456789, got %d", cfg.AppID)
	}
	if cfg.Mode != ModeGateway {
		t.Fatalf("expected mode gateway, got %v", cfg.Mode)
	}
	if cfg.Mode.String() != "gateway" {
		t.Fatalf("expected string gateway, got %s", cfg.Mode.String())
	}
	if ModeWebhook.String() != "webhook" {
		t.Fatalf("expected string webhook, got %s", ModeWebhook.String())
	}
	if Mode(99).String() != "unknown" {
		t.Fatalf("expected string unknown, got %s", Mode(99).String())
	}

	optMode := WithMode(ModeWebhook)
	optMode(&cfg)
	if cfg.Mode != ModeWebhook {
		t.Fatalf("expected mode webhook, got %v", cfg.Mode)
	}

	optIntents := WithIntents(types.IntentGuildMessages)
	optIntents(&cfg)
	if cfg.Intents != types.IntentGuildMessages {
		t.Fatalf("expected custom intents, got %v", cfg.Intents)
	}

	optDev := WithDevMode(true, 987654321)
	optDev(&cfg)
	if !cfg.DevMode || cfg.DevGuildID != 987654321 {
		t.Fatalf("expected dev mode enabled with guild id 987654321")
	}

	optDefer := WithAutoDefer(false, 3*time.Second)
	optDefer(&cfg)
	if cfg.AutoDefer {
		t.Fatalf("expected auto defer disabled")
	}
	if cfg.AutoDeferDelay != 3*time.Second {
		t.Fatalf("expected auto defer delay 3s, got %v", cfg.AutoDeferDelay)
	}

	optTimeout := WithTimeout(10 * time.Minute)
	optTimeout(&cfg)
	if cfg.Timeout != 10*time.Minute {
		t.Fatalf("expected timeout 10m, got %v", cfg.Timeout)
	}

	optWebhook := WithWebhook(":9090", "/custom-hook", "pubkey123")
	optWebhook(&cfg)
	if cfg.WebhookAddr != ":9090" || cfg.WebhookPath != "/custom-hook" || cfg.PublicKey != "pubkey123" {
		t.Fatalf("unexpected webhook config values")
	}
}

func TestConfig_Validation(t *testing.T) {
	cfg := Config{}
	if err := cfg.Validate(); !errors.Is(err, ErrTokenRequired) {
		t.Fatalf("expected ErrTokenRequired, got %v", err)
	}

	cfg.Token = "valid-token"
	if err := cfg.Validate(); !errors.Is(err, ErrAppIDRequired) {
		t.Fatalf("expected ErrAppIDRequired, got %v", err)
	}

	cfg.AppID = 123
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
}

func TestApp_NewAndConfig(t *testing.T) {
	_, err := New("", 0)
	if err == nil {
		t.Fatal("expected error for invalid config")
	}

	app, err := New("token", 12345, WithMode(ModeWebhook))
	if err != nil {
		t.Fatalf("unexpected error creating app: %v", err)
	}
	if app.Config().Mode != ModeWebhook {
		t.Fatalf("expected webhook mode in config, got %v", app.Config().Mode)
	}
	if app.REST == nil || app.Gateway == nil || app.Dispatcher == nil {
		t.Fatal("expected initialized REST, Gateway, and Dispatcher clients")
	}
}

func TestApp_CommandRegistration(t *testing.T) {
	app, err := New("token", 12345)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	invalidCmd := command.Command{}
	if err := app.RegisterCommand(invalidCmd, nil); err == nil {
		t.Fatal("expected validation error for empty command")
	}

	validCmd := command.Command{
		Type:        command.CommandTypeChatInput,
		Name:        "ping",
		Description: "ping test",
	}
	handler := func(c *context.Context) error { return nil }

	if err := app.RegisterCommand(validCmd, handler); err != nil {
		t.Fatalf("unexpected registration error: %v", err)
	}

	if err := app.RegisterCommand(validCmd, handler); !errors.Is(err, ErrDuplicateCommand) {
		t.Fatalf("expected ErrDuplicateCommand, got %v", err)
	}

	app2, _ := New("token", 12345)
	app2.Command(invalidCmd, handler)
	if app2.initErr == nil {
		t.Fatal("expected initErr on invalid command via fluent Command method")
	}

	app2.Command(validCmd, handler)
	if app2.lookupCommand(int(command.CommandTypeChatInput), validCmd.Name) != nil {
		t.Fatal("subsequent Command calls should be skipped if initErr is present")
	}
}

func TestApp_ComponentAndModalRegistration(t *testing.T) {
	app, _ := New("token", 12345)
	h := func(c *context.Context) error { return nil }

	app.Component("btn_ok", h)
	app.Modal("modal_feedback", h)

	if app.components["btn_ok"] == nil {
		t.Fatal("expected component handler registered")
	}
	if app.modals["modal_feedback"] == nil {
		t.Fatal("expected modal handler registered")
	}
}

func TestApp_Handle_NilContext(t *testing.T) {
	app, _ := New("token", 12345)
	if err := app.Handle(nil); err == nil {
		t.Fatal("expected error for nil context")
	}
	if err := app.Handle(&context.Context{}); err == nil {
		t.Fatal("expected error for context with nil interaction")
	}
}

func TestApp_Handle_Ping(t *testing.T) {
	app, _ := New("token", 12345)
	in := &types.Interaction{Type: 1}
	ctx := context.New(stdctx.Background(), in, nil)

	if err := app.Handle(ctx); err != nil {
		t.Fatalf("unexpected ping handle error: %v", err)
	}

	select {
	case resp := <-ctx.ResponseChannel():
		if resp.Type != types.CallbackTypePong {
			t.Fatalf("expected pong callback type, got %v", resp.Type)
		}
	default:
		t.Fatal("expected pong response in response channel")
	}
}

func TestApp_Handle_CommandPipeline(t *testing.T) {
	app, _ := New("token", 12345)
	var trace []string

	globalMw := func(next middleware.Handler) middleware.Handler {
		return func(c *context.Context) error {
			trace = append(trace, "global_before")
			err := next(c)
			trace = append(trace, "global_after")
			return err
		}
	}
	cmdMw := func(next middleware.Handler) middleware.Handler {
		return func(c *context.Context) error {
			trace = append(trace, "cmd_before")
			err := next(c)
			trace = append(trace, "cmd_after")
			return err
		}
	}

	app.Use(globalMw)
	cmd := command.Command{
		Type:        command.CommandTypeChatInput,
		Name:        "test",
		Description: "test command",
	}
	_ = app.RegisterCommand(cmd, func(c *context.Context) error {
		trace = append(trace, "handler")
		return nil
	}, cmdMw)

	in := &types.Interaction{
		Type: 2,
		Data: &types.InteractionData{
			Name: "test",
		},
	}
	ctx := context.New(stdctx.Background(), in, nil)

	if err := app.Handle(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []string{"global_before", "cmd_before", "handler", "cmd_after", "global_after"}
	if !reflect.DeepEqual(trace, expected) {
		t.Fatalf("expected trace %v, got %v", expected, trace)
	}
}

func TestApp_Handle_NotFoundErrors(t *testing.T) {
	app, _ := New("token", 12345)

	inCmdNilData := &types.Interaction{Type: 2, Data: nil}
	if err := app.Handle(context.New(stdctx.Background(), inCmdNilData, nil)); !errors.Is(err, ErrCommandNotFound) {
		t.Fatalf("expected ErrCommandNotFound, got %v", err)
	}

	inCmdNotFound := &types.Interaction{Type: 2, Data: &types.InteractionData{Name: "unknown"}}
	if err := app.Handle(context.New(stdctx.Background(), inCmdNotFound, nil)); !errors.Is(err, ErrCommandNotFound) {
		t.Fatalf("expected ErrCommandNotFound, got %v", err)
	}

	inCompNilData := &types.Interaction{Type: 3, Data: nil}
	if err := app.Handle(context.New(stdctx.Background(), inCompNilData, nil)); !errors.Is(err, ErrComponentNotFound) {
		t.Fatalf("expected ErrComponentNotFound, got %v", err)
	}

	inCompNotFound := &types.Interaction{Type: 3, Data: &types.InteractionData{CustomID: "unknown"}}
	if err := app.Handle(context.New(stdctx.Background(), inCompNotFound, nil)); !errors.Is(err, ErrComponentNotFound) {
		t.Fatalf("expected ErrComponentNotFound, got %v", err)
	}

	inModalNilData := &types.Interaction{Type: 5, Data: nil}
	if err := app.Handle(context.New(stdctx.Background(), inModalNilData, nil)); !errors.Is(err, ErrModalNotFound) {
		t.Fatalf("expected ErrModalNotFound, got %v", err)
	}

	inModalNotFound := &types.Interaction{Type: 5, Data: &types.InteractionData{CustomID: "unknown"}}
	if err := app.Handle(context.New(stdctx.Background(), inModalNotFound, nil)); !errors.Is(err, ErrModalNotFound) {
		t.Fatalf("expected ErrModalNotFound, got %v", err)
	}
}

func TestApp_Handle_ComponentAndModalSuccess(t *testing.T) {
	app, _ := New("token", 12345)

	compCalled := false
	app.Component("btn_click", func(c *context.Context) error {
		compCalled = true
		return nil
	})

	inComp := &types.Interaction{Type: 3, Data: &types.InteractionData{CustomID: "btn_click"}}
	if err := app.Handle(context.New(stdctx.Background(), inComp, nil)); err != nil {
		t.Fatalf("unexpected component error: %v", err)
	}
	if !compCalled {
		t.Fatal("expected component handler to be executed")
	}

	modalCalled := false
	app.Modal("modal_submit", func(c *context.Context) error {
		modalCalled = true
		return nil
	})

	inModal := &types.Interaction{Type: 5, Data: &types.InteractionData{CustomID: "modal_submit"}}
	if err := app.Handle(context.New(stdctx.Background(), inModal, nil)); err != nil {
		t.Fatalf("unexpected modal error: %v", err)
	}
	if !modalCalled {
		t.Fatal("expected modal handler to be executed")
	}
}

func TestApp_AutoDefer(t *testing.T) {
	app, _ := New("token", 12345, WithAutoDefer(true, 10*time.Millisecond))
	in := &types.Interaction{Type: 2, Data: &types.InteractionData{Name: "slow"}}

	cmd := command.Command{
		Type:        command.CommandTypeChatInput,
		Name:        "slow",
		Description: "slow desc",
	}
	_ = app.RegisterCommand(cmd, func(c *context.Context) error {
		time.Sleep(30 * time.Millisecond)
		return nil
	})

	ctx := context.New(stdctx.Background(), in, nil)
	_ = app.Handle(ctx)

	select {
	case resp := <-ctx.ResponseChannel():
		if resp.Type != types.CallbackTypeDeferredChannelMessageWithSource {
			t.Fatalf("expected deferred response type, got %v", resp.Type)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timed out waiting for auto defer")
	}
}

func TestApp_WebhookVerification(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate ed25519 key: %v", err)
	}
	pubHex := hex.EncodeToString(pub)

	app, err := New("token", 12345, WithWebhook(":0", "/interactions", pubHex), WithAutoDefer(false))
	if err != nil {
		t.Fatalf("unexpected app error: %v", err)
	}

	body := []byte(`{"type":1}`)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	msg := append([]byte(timestamp), body...)
	sig := ed25519.Sign(priv, msg)
	sigHex := hex.EncodeToString(sig)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mux := http.NewServeMux()
		mux.HandleFunc("/interactions", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			sHex := r.Header.Get("X-Signature-Ed25519")
			ts := r.Header.Get("X-Signature-Timestamp")
			if sHex == "" || ts == "" {
				http.Error(w, "invalid request signature", http.StatusUnauthorized)
				return
			}
			sec, err := strconv.ParseInt(ts, 10, 64)
			if err != nil {
				http.Error(w, "invalid request signature", http.StatusUnauthorized)
				return
			}
			diff := time.Since(time.Unix(sec, 0))
			if diff < -5*time.Minute || diff > 5*time.Minute {
				http.Error(w, "invalid request signature", http.StatusUnauthorized)
				return
			}
			sBytes, decErr := hex.DecodeString(sHex)
			if decErr != nil || len(sBytes) != ed25519.SignatureSize {
				http.Error(w, "invalid request signature", http.StatusUnauthorized)
				return
			}
			bBytes, _ := ioReadAll(r.Body)
			m := append([]byte(ts), bBytes...)
			if !ed25519.Verify(pub, m, sBytes) {
				http.Error(w, "invalid request signature", http.StatusUnauthorized)
				return
			}
			var in types.Interaction
			if decErr := json.Unmarshal(bBytes, &in); decErr != nil {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			gdkCtx := context.New(r.Context(), &in, app.REST)
			go func() {
				_ = app.Handle(gdkCtx)
			}()
			select {
			case resp := <-gdkCtx.ResponseChannel():
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(resp)
			case <-time.After(200 * time.Millisecond):
				w.WriteHeader(http.StatusOK)
			}
		})
		mux.ServeHTTP(w, r)
	})

	reqGet := httptest.NewRequest(http.MethodGet, "/interactions", nil)
	recGet := httptest.NewRecorder()
	handler.ServeHTTP(recGet, reqGet)
	if recGet.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 method not allowed, got %d", recGet.Code)
	}

	reqNoSig := httptest.NewRequest(http.MethodPost, "/interactions", bytes.NewReader(body))
	recNoSig := httptest.NewRecorder()
	handler.ServeHTTP(recNoSig, reqNoSig)
	if recNoSig.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthorized for missing signature, got %d", recNoSig.Code)
	}

	reqBadSig := httptest.NewRequest(http.MethodPost, "/interactions", bytes.NewReader(body))
	reqBadSig.Header.Set("X-Signature-Ed25519", "deadbeef")
	reqBadSig.Header.Set("X-Signature-Timestamp", timestamp)
	recBadSig := httptest.NewRecorder()
	handler.ServeHTTP(recBadSig, reqBadSig)
	if recBadSig.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 unauthorized for bad signature length, got %d", recBadSig.Code)
	}

	reqValid := httptest.NewRequest(http.MethodPost, "/interactions", bytes.NewReader(body))
	reqValid.Header.Set("X-Signature-Ed25519", sigHex)
	reqValid.Header.Set("X-Signature-Timestamp", timestamp)
	recValid := httptest.NewRecorder()
	handler.ServeHTTP(recValid, reqValid)
	if recValid.Code != http.StatusOK {
		t.Fatalf("expected 200 ok for valid ping, got %d", recValid.Code)
	}

	var resp types.InteractionResponse
	if err := json.Unmarshal(recValid.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Type != types.CallbackTypePong {
		t.Fatalf("expected pong response, got %v", resp.Type)
	}
}

func ioReadAll(r ioReader) ([]byte, error) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(r)
	return buf.Bytes(), err
}

type ioReader interface {
	Read(p []byte) (n int, err error)
}

func TestApp_RunAndClose(t *testing.T) {
	app, _ := New("token", 12345)
	app.initErr = errors.New("initialization failure")
	if err := app.Run(stdctx.Background()); err == nil {
		t.Fatal("expected error when initErr is present")
	}

	app2, _ := New("token", 12345)
	app2.running.Store(true)
	if err := app2.Run(stdctx.Background()); !errors.Is(err, ErrAlreadyRunning) {
		t.Fatalf("expected ErrAlreadyRunning, got %v", err)
	}

	app3, _ := New("token", 12345, WithMode(ModeWebhook))
	if err := app3.Run(stdctx.Background()); !errors.Is(err, ErrPublicKeyRequired) {
		t.Fatalf("expected ErrPublicKeyRequired, got %v", err)
	}

	app4, _ := New("token", 12345, WithWebhook(":0", "/hook", "invalid-hex"))
	if err := app4.Run(stdctx.Background()); !errors.Is(err, ErrInvalidSignature) {
		t.Fatalf("expected ErrInvalidSignature for invalid hex public key, got %v", err)
	}

	if err := app2.Close(); err != nil {
		t.Fatalf("unexpected error closing app: %v", err)
	}
}
