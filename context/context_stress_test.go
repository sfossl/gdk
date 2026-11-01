package context

import (
	stdctx "context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

func TestStress_ConcurrentRespondInitialAndFollowups(t *testing.T) {
	var initialCalls int64
	var followupCalls int64

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/interactions/100/test-token/callback":
			atomic.AddInt64(&initialCalls, 1)
			w.WriteHeader(http.StatusNoContent)
		case "/webhooks/50/test-token":
			atomic.AddInt64(&followupCalls, 1)
			_ = json.NewEncoder(w).Encode(types.Message{
				ID: types.Snowflake(1),
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))
	in := &types.Interaction{
		ID:    types.Snowflake(100),
		AppID: types.Snowflake(50),
		Token: "test-token",
	}
	ctx := NewGatewayContext(stdctx.Background(), in, client)

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)

	start := make(chan struct{})

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			<-start
			_, err := ctx.Respond(Response{
				Content: "concurrent test message",
			})
			if err != nil {
				t.Errorf("unexpected error on respond: %v", err)
			}
		}(i)
	}

	close(start)
	wg.Wait()

	if initialCalls != 1 {
		t.Fatalf("expected exactly 1 initial interaction callback, got %d", initialCalls)
	}
	if followupCalls != goroutines-1 {
		t.Fatalf("expected %d followup messages, got %d", goroutines-1, followupCalls)
	}
}

func TestStress_OptionValCoercionAndResolvedSynthesis(t *testing.T) {
	targetUID := types.Snowflake(777)
	resolved := &types.ResolvedData{
		Users: map[types.Snowflake]types.User{
			targetUID: {
				ID:       targetUID,
				Username: "resolved_user",
			},
		},
		Members: map[types.Snowflake]types.Member{
			targetUID: {
				Roles: []types.Snowflake{1, 2, 3},
			},
		},
	}

	in := &types.Interaction{
		Data: &types.InteractionData{
			Resolved: resolved,
			Options: []types.ApplicationCommandOption{
				{
					Name:  "str_num",
					Type:  3,
					Value: "42",
				},
				{
					Name:  "str_bool",
					Type:  3,
					Value: "true",
				},
				{
					Name:  "str_float",
					Type:  3,
					Value: "3.1415",
				},
				{
					Name:  "str_snowflake",
					Type:  3,
					Value: "777",
				},
			},
		},
	}

	ctx := New(stdctx.Background(), in, nil)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			if val := ctx.Opt("str_num").Int(); val != 42 {
				t.Errorf("expected 42 from string, got %d", val)
			}
			if val := ctx.Opt("str_num").Int64(); val != 42 {
				t.Errorf("expected 42 from string, got %d", val)
			}
			if val := ctx.Opt("str_bool").Bool(); !val {
				t.Errorf("expected true from string, got %v", val)
			}
			if val := ctx.Opt("str_float").Float(); val < 3.14 || val > 3.15 {
				t.Errorf("expected 3.1415 from string, got %f", val)
			}

			optUser := ctx.Opt("str_snowflake")
			user := optUser.User()
			if user == nil || user.Username != "resolved_user" {
				t.Errorf("failed to resolve user: %+v", user)
			}

			member := optUser.Member()
			if member == nil {
				t.Errorf("failed to resolve member")
			} else if member.User == nil || member.User.Username != "resolved_user" {
				t.Errorf("failed to synthesize member.User from resolved users: %+v", member)
			}
		}()
	}
	wg.Wait()
}

func TestStress_HierarchyBoundaryConditions(t *testing.T) {
	guildID := types.Snowflake(100)
	ownerID := types.Snowflake(999)
	adminID := types.Snowflake(888)
	botID := types.Snowflake(50)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/guilds/100":
			_ = json.NewEncoder(w).Encode(types.Guild{
				ID:      guildID,
				OwnerID: ownerID,
				Roles: []types.Role{
					{ID: 10, Position: 10},
					{ID: 20, Position: 20},
					{ID: 30, Position: 30},
				},
			})
		case "/guilds/100/members/50":
			_ = json.NewEncoder(w).Encode(types.Member{
				Roles: []types.Snowflake{20},
			})
		case "/guilds/100/members/777":
			_ = json.NewEncoder(w).Encode(types.Member{
				Roles: []types.Snowflake{20},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))

	ctxOwner := &Context{
		ctx:     stdctx.Background(),
		AppID:   botID,
		GuildID: &guildID,
		REST:    client,
		Author:  &types.User{ID: ownerID},
	}
	ok, err := ctxOwner.CanModerate(ownerID)
	if ok || !errors.Is(err, ErrCannotModerateOwner) {
		t.Fatalf("expected ErrCannotModerateOwner, got ok=%v, err=%v", ok, err)
	}

	ctxSelf := &Context{
		ctx:     stdctx.Background(),
		AppID:   botID,
		GuildID: &guildID,
		REST:    client,
		Author:  &types.User{ID: adminID},
	}
	ok, err = ctxSelf.CanModerate(adminID)
	if ok || !errors.Is(err, ErrCannotModerateSelf) {
		t.Fatalf("expected ErrCannotModerateSelf, got ok=%v, err=%v", ok, err)
	}

	ctxEqualRank := &Context{
		ctx:     stdctx.Background(),
		AppID:   botID,
		GuildID: &guildID,
		REST:    client,
		Author:  &types.User{ID: adminID},
		InvokerMember: &types.Member{
			Roles: []types.Snowflake{30},
		},
	}
	ok, err = ctxEqualRank.CanModerate(777)
	if ok || !errors.Is(err, ErrTargetOutranksBot) {
		t.Fatalf("expected ErrTargetOutranksBot when bot pos 20 <= target pos 20, got ok=%v, err=%v", ok, err)
	}
}

func TestStress_SlowmodeAndModerationLimits(t *testing.T) {
	if _, err := parseSlowmodeSeconds(-1); !errors.Is(err, ErrSlowmodeNegative) {
		t.Fatalf("expected ErrSlowmodeNegative, got %v", err)
	}
	if _, err := parseSlowmodeSeconds(21601); !errors.Is(err, ErrSlowmodeTooLong) {
		t.Fatalf("expected ErrSlowmodeTooLong, got %v", err)
	}
	if _, err := parseSlowmodeSeconds("21601s"); !errors.Is(err, ErrSlowmodeTooLong) {
		t.Fatalf("expected ErrSlowmodeTooLong from duration string, got %v", err)
	}

	if _, err := parseTimeoutUntil(time.Duration(0)); !errors.Is(err, ErrTimeoutTooShort) {
		t.Fatalf("expected ErrTimeoutTooShort, got %v", err)
	}
	if _, err := parseTimeoutUntil(29 * 24 * time.Hour); !errors.Is(err, ErrTimeoutTooLong) {
		t.Fatalf("expected ErrTimeoutTooLong, got %v", err)
	}

	if _, err := parseDeleteHistorySeconds(MaxBanDeleteSec + 1); !errors.Is(err, ErrBanDeleteTooLong) {
		t.Fatalf("expected ErrBanDeleteTooLong, got %v", err)
	}
}
