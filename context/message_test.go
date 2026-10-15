package context

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sfossl/gda/rest"
	"github.com/sfossl/gda/types"
)

func setupMessageTestContext(handler http.HandlerFunc) (*Context, func()) {
	ts := httptest.NewServer(handler)
	client := rest.NewClient("test-token", rest.WithBaseURL(ts.URL))

	c := &Context{
		ctx:       context.Background(),
		ChannelID: types.Snowflake(200),
		REST:      client,
	}
	return c, ts.Close
}

func TestMessageDefaultsAndTargeting(t *testing.T) {
	targetID := types.Snowflake(777)
	c := &Context{
		ChannelID: types.Snowflake(200),
		Interaction: &types.Interaction{
			Data: &types.InteractionData{
				TargetID: &targetID,
			},
		},
	}

	act := c.Message(types.Snowflake(500))
	if act.ID() != types.Snowflake(500) {
		t.Fatalf("expected message id 500, got %v", act.ID())
	}
	if act.ChannelID() != types.Snowflake(200) {
		t.Fatalf("expected channel id 200, got %v", act.ChannelID())
	}

	actCustom := c.Message(types.Snowflake(500), types.Snowflake(300))
	if actCustom.ChannelID() != types.Snowflake(300) {
		t.Fatalf("expected channel id 300, got %v", actCustom.ChannelID())
	}

	actManage := c.ManageMessage(types.Snowflake(600))
	if actManage.ID() != types.Snowflake(600) {
		t.Fatalf("expected message id 600, got %v", actManage.ID())
	}

	actTarget := c.TargetMessage()
	if actTarget.ID() != types.Snowflake(777) {
		t.Fatalf("expected target id 777, got %v", actTarget.ID())
	}
}

func TestMessageValidationErrors(t *testing.T) {
	cNoREST := &Context{ChannelID: types.Snowflake(200)}
	actNoREST := cNoREST.Message(types.Snowflake(500))

	if _, err := actNoREST.Get(); !errors.Is(err, ErrNoRESTClient) {
		t.Fatalf("expected ErrNoRESTClient, got %v", err)
	}
	if _, err := actNoREST.Edit(rest.MessagePayload{}); !errors.Is(err, ErrNoRESTClient) {
		t.Fatalf("expected ErrNoRESTClient, got %v", err)
	}
	if err := actNoREST.Delete(); !errors.Is(err, ErrNoRESTClient) {
		t.Fatalf("expected ErrNoRESTClient, got %v", err)
	}
	if err := actNoREST.Pin(); !errors.Is(err, ErrNoRESTClient) {
		t.Fatalf("expected ErrNoRESTClient, got %v", err)
	}
	if err := actNoREST.Unpin(); !errors.Is(err, ErrNoRESTClient) {
		t.Fatalf("expected ErrNoRESTClient, got %v", err)
	}
	if err := actNoREST.React("star"); !errors.Is(err, ErrNoRESTClient) {
		t.Fatalf("expected ErrNoRESTClient, got %v", err)
	}
	if err := actNoREST.RemoveReaction("star"); !errors.Is(err, ErrNoRESTClient) {
		t.Fatalf("expected ErrNoRESTClient, got %v", err)
	}

	c, cleanup := setupMessageTestContext(func(w http.ResponseWriter, r *http.Request) {})
	defer cleanup()

	actZeroMsg := c.Message(0)
	if _, err := actZeroMsg.Get(); !errors.Is(err, ErrInvalidMessageID) {
		t.Fatalf("expected ErrInvalidMessageID, got %v", err)
	}
	if _, err := actZeroMsg.Edit(rest.MessagePayload{}); !errors.Is(err, ErrInvalidMessageID) {
		t.Fatalf("expected ErrInvalidMessageID, got %v", err)
	}
	if err := actZeroMsg.Delete(); !errors.Is(err, ErrInvalidMessageID) {
		t.Fatalf("expected ErrInvalidMessageID, got %v", err)
	}
	if err := actZeroMsg.Pin(); !errors.Is(err, ErrInvalidMessageID) {
		t.Fatalf("expected ErrInvalidMessageID, got %v", err)
	}
	if err := actZeroMsg.Unpin(); !errors.Is(err, ErrInvalidMessageID) {
		t.Fatalf("expected ErrInvalidMessageID, got %v", err)
	}
	if err := actZeroMsg.React("star"); !errors.Is(err, ErrInvalidMessageID) {
		t.Fatalf("expected ErrInvalidMessageID, got %v", err)
	}
	if err := actZeroMsg.RemoveReaction("star"); !errors.Is(err, ErrInvalidMessageID) {
		t.Fatalf("expected ErrInvalidMessageID, got %v", err)
	}

	actZeroCh := c.Message(types.Snowflake(500))
	actZeroCh.channelID = 0
	if _, err := actZeroCh.Get(); !errors.Is(err, ErrInvalidChannelID) {
		t.Fatalf("expected ErrInvalidChannelID, got %v", err)
	}
	if _, err := actZeroCh.Edit(rest.MessagePayload{}); !errors.Is(err, ErrInvalidChannelID) {
		t.Fatalf("expected ErrInvalidChannelID, got %v", err)
	}
	if err := actZeroCh.Delete(); !errors.Is(err, ErrInvalidChannelID) {
		t.Fatalf("expected ErrInvalidChannelID, got %v", err)
	}
	if err := actZeroCh.Pin(); !errors.Is(err, ErrInvalidChannelID) {
		t.Fatalf("expected ErrInvalidChannelID, got %v", err)
	}
	if err := actZeroCh.Unpin(); !errors.Is(err, ErrInvalidChannelID) {
		t.Fatalf("expected ErrInvalidChannelID, got %v", err)
	}
	if err := actZeroCh.React("star"); !errors.Is(err, ErrInvalidChannelID) {
		t.Fatalf("expected ErrInvalidChannelID, got %v", err)
	}
	if err := actZeroCh.RemoveReaction("star"); !errors.Is(err, ErrInvalidChannelID) {
		t.Fatalf("expected ErrInvalidChannelID, got %v", err)
	}

	actValid := c.Message(types.Snowflake(500))
	if err := actValid.React(""); !errors.Is(err, ErrEmojiRequired) {
		t.Fatalf("expected ErrEmojiRequired, got %v", err)
	}
	if err := actValid.RemoveReaction("  "); !errors.Is(err, ErrEmojiRequired) {
		t.Fatalf("expected ErrEmojiRequired, got %v", err)
	}
}

func TestMessageGetAndEdit(t *testing.T) {
	c, cleanup := setupMessageTestContext(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && r.URL.Path == "/channels/200/messages/500" {
			_ = json.NewEncoder(w).Encode(types.Message{
				ID:        types.Snowflake(500),
				ChannelID: types.Snowflake(200),
				Content:   "hello",
			})
			return
		}
		if r.Method == http.MethodPatch && r.URL.Path == "/channels/200/messages/500" {
			var p rest.MessagePayload
			_ = json.NewDecoder(r.Body).Decode(&p)
			_ = json.NewEncoder(w).Encode(types.Message{
				ID:         types.Snowflake(500),
				ChannelID:  types.Snowflake(200),
				Content:    p.Content,
				Embeds:     p.Embeds,
				Components: p.Components,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer cleanup()

	act := c.Message(types.Snowflake(500))

	msg, err := act.Get()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if msg.Content != "hello" {
		t.Fatalf("expected hello, got %s", msg.Content)
	}

	mContent, err := act.SetContent("updated text")
	if err != nil || mContent.Content != "updated text" {
		t.Fatalf("expected updated text, got %v err %v", mContent, err)
	}

	mEmbed, err := act.SetEmbeds(types.Embed{Title: "embed title"})
	if err != nil || len(mEmbed.Embeds) != 1 || mEmbed.Embeds[0].Title != "embed title" {
		t.Fatalf("expected embed title, got %v err %v", mEmbed, err)
	}

	mComp, err := act.SetComponents(types.Component{Type: 1})
	if err != nil || len(mComp.Components) != 1 {
		t.Fatalf("expected component, got %v err %v", mComp, err)
	}
}

func TestMessageDeleteAndPins(t *testing.T) {
	c, cleanup := setupMessageTestContext(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == "/channels/200/messages/500" {
			if r.Header.Get("X-Audit-Log-Reason") != "spam purge" {
				t.Fatalf("unexpected audit header %s", r.Header.Get("X-Audit-Log-Reason"))
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodPut && r.URL.Path == "/channels/200/pins/500" {
			if r.Header.Get("X-Audit-Log-Reason") != "pinned by mod" {
				t.Fatalf("unexpected audit header %s", r.Header.Get("X-Audit-Log-Reason"))
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodDelete && r.URL.Path == "/channels/200/pins/500" {
			if r.Header.Get("X-Audit-Log-Reason") != "unpinned" {
				t.Fatalf("unexpected audit header %s", r.Header.Get("X-Audit-Log-Reason"))
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer cleanup()

	act := c.Message(types.Snowflake(500))

	if err := act.Delete("spam", "purge"); err != nil {
		t.Fatalf("unexpected error deleting message: %v", err)
	}
	if err := act.Pin("pinned", "by", "mod"); err != nil {
		t.Fatalf("unexpected error pinning message: %v", err)
	}
	if err := act.Unpin("unpinned"); err != nil {
		t.Fatalf("unexpected error unpinning message: %v", err)
	}
}

func TestMessageReactions(t *testing.T) {
	c, cleanup := setupMessageTestContext(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut && r.URL.Path == "/channels/200/messages/500/reactions/thumbsup/@me" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodDelete && r.URL.Path == "/channels/200/messages/500/reactions/thumbsup/@me" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method == http.MethodDelete && r.URL.Path == "/channels/200/messages/500/reactions/thumbsup/999" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	})
	defer cleanup()

	act := c.Message(types.Snowflake(500))

	if err := act.React("thumbsup"); err != nil {
		t.Fatalf("unexpected error adding reaction: %v", err)
	}
	if err := act.RemoveReaction("thumbsup"); err != nil {
		t.Fatalf("unexpected error removing own reaction: %v", err)
	}
	if err := act.RemoveReaction("thumbsup", types.Snowflake(999)); err != nil {
		t.Fatalf("unexpected error removing user reaction: %v", err)
	}
}
