package app

import (
	"bytes"
	stdctx "context"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/command"
	"github.com/sfossl/gdk/context"
)

func TestStress_WebhookTamperingAndVerification(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("failed to generate ed25519 key: %v", err)
	}

	pubHex := hex.EncodeToString(pub)

	app, err := New("test-token", types.Snowflake(1),
		WithWebhook("127.0.0.1:0", "/interactions", pubHex),
		WithAutoDefer(false),
	)
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	app.Command(command.Command{Name: "ping", Description: "ping cmd"}, func(ctx *context.Context) error {
		_, err := ctx.Respond(context.Response{Content: "pong"})
		return err
	})

	inBytes, _ := json.Marshal(types.Interaction{
		ID:    100,
		Token: "tok",
		Type:  2,
		Data: &types.InteractionData{
			Name: "ping",
		},
	})

	validTimestamp := strconv.FormatInt(time.Now().Unix(), 10)
	msg := append([]byte(validTimestamp), inBytes...)
	validSig := hex.EncodeToString(ed25519.Sign(priv, msg))

	corruptSig := hex.EncodeToString(make([]byte, ed25519.SignatureSize))

	var serverHandler http.Handler
	app.server = &http.Server{}

	mux := http.NewServeMux()
	mux.HandleFunc(app.config.WebhookPath, func(w http.ResponseWriter, r *http.Request) {
		sigHex := r.Header.Get("X-Signature-Ed25519")
		timestamp := r.Header.Get("X-Signature-Timestamp")
		if sigHex == "" || timestamp == "" {
			http.Error(w, "invalid request signature", http.StatusUnauthorized)
			return
		}
		sec, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			http.Error(w, "invalid request signature", http.StatusUnauthorized)
			return
		}
		diff := time.Since(time.Unix(sec, 0))
		if diff < -5*time.Minute || diff > 5*time.Minute {
			http.Error(w, "invalid request signature", http.StatusUnauthorized)
			return
		}
		sig, err := hex.DecodeString(sigHex)
		if err != nil || len(sig) != ed25519.SignatureSize {
			http.Error(w, "invalid request signature", http.StatusUnauthorized)
			return
		}
		body := make([]byte, r.ContentLength)
		_, _ = r.Body.Read(body)
		vmsg := append([]byte(timestamp), body...)
		if !ed25519.Verify(pub, vmsg, sig) {
			http.Error(w, "invalid request signature", http.StatusUnauthorized)
			return
		}
		var in types.Interaction
		_ = json.Unmarshal(body, &in)
		gdkCtx := context.New(stdctx.Background(), &in, app.REST)
		_ = app.Handle(gdkCtx)
		select {
		case resp := <-gdkCtx.ResponseChannel():
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		case <-time.After(100 * time.Millisecond):
		}
	})
	serverHandler = mux

	ts := httptest.NewServer(serverHandler)
	defer ts.Close()

	const iterations = 50
	var wg sync.WaitGroup

	for i := 0; i < iterations; i++ {
		wg.Add(4)

		go func() {
			defer wg.Done()
			req, _ := http.NewRequest(http.MethodPost, ts.URL+"/interactions", bytes.NewReader(inBytes))
			req.Header.Set("X-Signature-Ed25519", validSig)
			req.Header.Set("X-Signature-Timestamp", validTimestamp)
			resp, err := http.DefaultClient.Do(req)
			if err != nil || resp.StatusCode != http.StatusOK {
				t.Errorf("expected 200 for valid signed request, got: %v", resp.StatusCode)
			}
		}()

		go func() {
			defer wg.Done()
			req, _ := http.NewRequest(http.MethodPost, ts.URL+"/interactions", bytes.NewReader(inBytes))
			req.Header.Set("X-Signature-Ed25519", corruptSig)
			req.Header.Set("X-Signature-Timestamp", validTimestamp)
			resp, err := http.DefaultClient.Do(req)
			if err != nil || resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("expected 401 for corrupted signature, got: %v", resp.StatusCode)
			}
		}()

		go func() {
			defer wg.Done()
			tamperedBody := append(inBytes, []byte("tamper")...)
			req, _ := http.NewRequest(http.MethodPost, ts.URL+"/interactions", bytes.NewReader(tamperedBody))
			req.Header.Set("X-Signature-Ed25519", validSig)
			req.Header.Set("X-Signature-Timestamp", validTimestamp)
			resp, err := http.DefaultClient.Do(req)
			if err != nil || resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("expected 401 for tampered body, got: %v", resp.StatusCode)
			}
		}()

		go func() {
			defer wg.Done()
			req, _ := http.NewRequest(http.MethodPost, ts.URL+"/interactions", bytes.NewReader(inBytes))
			req.Header.Set("X-Signature-Ed25519", validSig)
			req.Header.Set("X-Signature-Timestamp", "1600000000")
			resp, err := http.DefaultClient.Do(req)
			if err != nil || resp.StatusCode != http.StatusUnauthorized {
				t.Errorf("expected 401 for stale timestamp replay, got: %v", resp.StatusCode)
			}
		}()
	}

	wg.Wait()
}

func TestStress_SyncHashDeterminismAndOrdering(t *testing.T) {
	cmds1 := []types.ApplicationCommand{
		{Name: "zebra", Description: "z", Type: 1},
		{Name: "alpha", Description: "a", Type: 1},
		{Name: "beta", Description: "b", Type: 1},
		{Name: "alpha", Description: "a user ctx", Type: 2},
	}

	cmds2 := []types.ApplicationCommand{
		{Name: "alpha", Description: "a user ctx", Type: 2},
		{Name: "beta", Description: "b", Type: 1},
		{Name: "zebra", Description: "z", Type: 1},
		{Name: "alpha", Description: "a", Type: 1},
	}

	sortCommands(cmds1)
	sortCommands(cmds2)

	h1, err1 := hashCommands(cmds1)
	if err1 != nil {
		t.Fatalf("unexpected hash error 1: %v", err1)
	}

	h2, err2 := hashCommands(cmds2)
	if err2 != nil {
		t.Fatalf("unexpected hash error 2: %v", err2)
	}

	if h1 != h2 {
		t.Fatalf("expected identical deterministic hashes, got %s vs %s", h1, h2)
	}
}

func TestStress_AppBuilderConcurrency(t *testing.T) {
	app, err := New("token", types.Snowflake(1))
	if err != nil {
		t.Fatalf("unexpected error creating app: %v", err)
	}

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 3)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			cmdName := fmt.Sprintf("cmd_%d", id)
			_ = app.RegisterCommand(command.Command{
				Name:        cmdName,
				Description: "test command",
			}, func(ctx *context.Context) error {
				return nil
			})
		}(i)

		go func(id int) {
			defer wg.Done()
			app.Component(fmt.Sprintf("btn_%d", id), func(ctx *context.Context) error {
				return nil
			})
		}(i)

		go func(id int) {
			defer wg.Done()
			app.Modal(fmt.Sprintf("modal_%d", id), func(ctx *context.Context) error {
				return nil
			})
		}(i)
	}

	wg.Wait()

	app.mu.RLock()
	cmdCount := len(app.commands)
	compCount := len(app.components)
	modalCount := len(app.modals)
	app.mu.RUnlock()

	if cmdCount != goroutines {
		t.Fatalf("expected %d commands, got %d", goroutines, cmdCount)
	}
	if compCount != goroutines {
		t.Fatalf("expected %d components, got %d", goroutines, compCount)
	}
	if modalCount != goroutines {
		t.Fatalf("expected %d modals, got %d", goroutines, modalCount)
	}
}

func TestStress_SyncFullCycle(t *testing.T) {
	var bulkGlobalCalls int64
	var bulkGuildCalls int64

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/applications/1/commands":
			atomic.AddInt64(&bulkGlobalCalls, 1)
			_ = json.NewEncoder(w).Encode([]types.ApplicationCommand{})
		case "/applications/1/guilds/777/commands":
			atomic.AddInt64(&bulkGuildCalls, 1)
			_ = json.NewEncoder(w).Encode([]types.ApplicationCommand{})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))

	app, err := New("test-token", types.Snowflake(1))
	if err != nil {
		t.Fatalf("unexpected app creation error: %v", err)
	}
	app.REST = client

	_ = app.RegisterCommand(command.Command{
		Name:        "global1",
		Description: "global command",
	}, nil)

	_ = app.RegisterCommand(command.Command{
		Name:        "guild1",
		Description: "guild command",
		GuildID:     777,
	}, nil)

	res1, err := app.SyncWithResult(stdctx.Background())
	if err != nil {
		t.Fatalf("first sync failed: %v", err)
	}
	if !res1.GlobalSynced || !res1.GuildsSynced[777] || res1.Updated != 2 || res1.Skipped != 0 {
		t.Fatalf("unexpected first sync result: %+v", res1)
	}

	res2, err := app.SyncWithResult(stdctx.Background())
	if err != nil {
		t.Fatalf("second sync failed: %v", err)
	}
	if res2.Updated != 0 || res2.Skipped != 2 {
		t.Fatalf("expected all skipped on unchanged sync, got: %+v", res2)
	}

	res3, err := app.SyncWithResult(stdctx.Background(), true)
	if err != nil {
		t.Fatalf("force sync failed: %v", err)
	}
	if res3.Updated != 2 || res3.Skipped != 0 {
		t.Fatalf("expected all updated on force sync, got: %+v", res3)
	}
}

func TestStress_AppErrAndLifecycle(t *testing.T) {
	app, _ := New("token", types.Snowflake(1))
	if app.Err() != nil {
		t.Fatalf("expected nil Err initially, got: %v", app.Err())
	}

	app.Command(command.Command{Name: "INVALID NAME UPPERCASE", Description: "invalid"}, nil)
	if app.Err() == nil {
		t.Fatal("expected Err after invalid command registration")
	}

	err := app.Run(stdctx.Background())
	if !errors.Is(err, app.Err()) {
		t.Fatalf("expected Run to return initErr %v, got %v", app.Err(), err)
	}
}

func TestStress_AutocompleteDispatch(t *testing.T) {
	app, _ := New("token", types.Snowflake(1))

	var handled bool
	app.Command(command.Command{
		Name:        "search",
		Description: "search query",
		AutocompleteHandler: func(ctx *context.Context) error {
			handled = true
			return ctx.Autocomplete([]types.Choice{
				{Name: "Option A", Value: "opt_a"},
				{Name: "Option B", Value: "opt_b"},
			})
		},
	}, nil)

	in := &types.Interaction{
		ID:    1,
		Token: "tok",
		Type:  4,
		Data: &types.InteractionData{
			Name: "search",
		},
	}
	ctx := context.New(stdctx.Background(), in, nil)
	if err := app.Handle(ctx); err != nil {
		t.Fatalf("unexpected autocomplete handle error: %v", err)
	}
	if !handled {
		t.Fatal("expected autocomplete handler to be called")
	}

	select {
	case resp := <-ctx.ResponseChannel():
		if resp.Type != types.CallbackTypeApplicationCommandAutocompleteResult {
			t.Fatalf("expected callback type 8, got %v", resp.Type)
		}
		if len(resp.Data.Choices) != 2 {
			t.Fatalf("expected 2 choices, got %d", len(resp.Data.Choices))
		}
	default:
		t.Fatal("expected autocomplete response on channel")
	}
}


