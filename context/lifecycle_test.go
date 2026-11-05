package context

import (
	stdctx "context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

func TestDisableAllComponents(t *testing.T) {
	btn1 := types.Component{Type: types.ComponentTypeButton, CustomID: "b1", Disabled: false}
	btn2 := types.Component{Type: types.ComponentTypeButton, CustomID: "b2", Disabled: false}
	row := types.Component{
		Type:       types.ComponentTypeActionRow,
		Components: []types.Component{btn1, btn2},
	}

	disabled := DisableAllComponents([]types.Component{row})
	if len(disabled) != 1 {
		t.Fatalf("expected 1 row, got %d", len(disabled))
	}
	if len(disabled[0].Components) != 2 {
		t.Fatalf("expected 2 buttons, got %d", len(disabled[0].Components))
	}
	if !disabled[0].Components[0].Disabled || !disabled[0].Components[1].Disabled {
		t.Fatalf("expected all buttons to be disabled")
	}
}

func TestResolveOptions(t *testing.T) {
	opts := Options{
		ComponentTimeout: 30 * time.Second,
		DeleteAfter:      "15s",
	}

	if opts.ResolveComponentTimeout() != 30*time.Second {
		t.Fatalf("expected 30s, got %v", opts.ResolveComponentTimeout())
	}
	if opts.ResolveDeleteAfter() != 15*time.Second {
		t.Fatalf("expected 15s, got %v", opts.ResolveDeleteAfter())
	}

	optsCompound := Options{
		ComponentTimeout: "1d",
		DeleteAfter:      "1h 30m",
	}
	if optsCompound.ResolveComponentTimeout() != 24*time.Hour {
		t.Fatalf("expected 24h for 1d, got %v", optsCompound.ResolveComponentTimeout())
	}
	if optsCompound.ResolveDeleteAfter() != 90*time.Minute {
		t.Fatalf("expected 90m for 1h 30m, got %v", optsCompound.ResolveDeleteAfter())
	}
}

func TestLifecycle_ConflictResolution(t *testing.T) {
	var editCalls atomic.Int32
	var deleteCalls atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			deleteCalls.Add(1)
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodPatch {
			editCalls.Add(1)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(types.Message{ID: 123})
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	restClient := rest.NewClient("token", rest.WithBaseURL(server.URL))
	in := &types.Interaction{
		ID:    111,
		AppID: 222,
		Token: "test_token",
	}

	ctx := New(stdctx.Background(), in, restClient)

	btn := types.Component{Type: types.ComponentTypeButton, CustomID: "test_btn"}
	row := types.Component{Type: types.ComponentTypeActionRow, Components: []types.Component{btn}}

	ctx.scheduleLifecycleTimers(Options{
		Components:       []types.Component{row},
		DeleteAfter:      10 * time.Millisecond,
		ComponentTimeout: 50 * time.Millisecond,
	}, nil)

	time.Sleep(100 * time.Millisecond)

	if deleteCalls.Load() != 1 {
		t.Fatalf("expected exactly 1 delete call, got %d", deleteCalls.Load())
	}
	if editCalls.Load() != 0 {
		t.Fatalf("expected 0 edit calls because deletion canceled component timeout, got %d", editCalls.Load())
	}
}

func TestLifecycle_DisableComponents(t *testing.T) {
	btn := types.Component{Type: types.ComponentTypeButton, CustomID: "b1"}
	row := types.Component{Type: types.ComponentTypeActionRow, Components: []types.Component{btn}}

	in := &types.Interaction{
		ID:    111,
		AppID: 222,
		Token: "token",
		Message: &types.Message{
			ID:         333,
			Components: []types.Component{row},
		},
	}

	ctx := New(stdctx.Background(), in, nil)
	err := ctx.DisableComponents()
	if err != nil {
		t.Fatalf("unexpected error disabling components: %v", err)
	}

	select {
	case res := <-ctx.ResponseChannel():
		if res.Type != types.CallbackTypeUpdateMessage {
			t.Fatalf("expected callback type %d, got %d", types.CallbackTypeUpdateMessage, res.Type)
		}
		if len(res.Data.Components) == 0 || len(res.Data.Components[0].Components) == 0 || !res.Data.Components[0].Components[0].Disabled {
			t.Fatal("expected button component to be disabled")
		}
	default:
		t.Fatal("expected disable call to produce update message on channel")
	}
}
