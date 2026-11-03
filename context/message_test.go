package context

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

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

	act := c.MessageSuite(types.Snowflake(500))
	if act.ID() != types.Snowflake(500) {
		t.Fatalf("expected message id 500, got %v", act.ID())
	}
	if act.ChannelID() != types.Snowflake(200) {
		t.Fatalf("expected channel id 200, got %v", act.ChannelID())
	}

	actCustom := c.MessageSuite(types.Snowflake(500), types.Snowflake(300))
	if actCustom.ChannelID() != types.Snowflake(300) {
		t.Fatalf("expected channel id 300, got %v", actCustom.ChannelID())
	}

	actTarget := c.TargetMessage()
	if actTarget.ID() != types.Snowflake(777) {
		t.Fatalf("expected target id 777, got %v", actTarget.ID())
	}
}

func TestManageMessage(t *testing.T) {
	cNoREST := &Context{ChannelID: types.Snowflake(200)}
	if err := cNoREST.ManageMessage(MessageOptions{Type: MsgPin, MessageID: 100}); !errors.Is(err, ErrNoRESTClient) {
		t.Fatalf("expected ErrNoRESTClient, got %v", err)
	}

	c, cleanup := setupMessageTestContext(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/channels/200/pins/100":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/channels/200/pins/100":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/channels/200/messages/100":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == "/channels/300/messages/100/crosspost":
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer cleanup()

	if err := c.ManageMessage(MessageOptions{Type: MsgPin, MessageID: 0}); !errors.Is(err, ErrInvalidMessageID) {
		t.Fatalf("expected ErrInvalidMessageID, got %v", err)
	}

	cNoCh := &Context{REST: c.REST}
	if err := cNoCh.ManageMessage(MessageOptions{Type: MsgPin, MessageID: 100}); !errors.Is(err, ErrInvalidChannelID) {
		t.Fatalf("expected ErrInvalidChannelID, got %v", err)
	}

	if err := c.ManageMessage(MessageOptions{Type: MessageAction(999), MessageID: 100}); !errors.Is(err, ErrInvalidMessageAction) {
		t.Fatalf("expected ErrInvalidMessageAction, got %v", err)
	}

	if err := c.ManageMessage(MessageOptions{Type: MsgPin, MessageID: 100}); err != nil {
		t.Fatalf("unexpected error pinning: %v", err)
	}

	if err := c.ManageMessage(MessageOptions{Type: MsgUnpin, MessageID: 100}); err != nil {
		t.Fatalf("unexpected error unpinning: %v", err)
	}

	if err := c.ManageMessage(MessageOptions{Type: MsgDelete, MessageID: 100}); err != nil {
		t.Fatalf("unexpected error deleting: %v", err)
	}

	if err := c.ManageMessage(MessageOptions{Type: MsgCrosspost, MessageID: 100, ChannelID: 300}); err != nil {
		t.Fatalf("unexpected error crossposting: %v", err)
	}
}

func TestMessageValidationErrors(t *testing.T) {
	cNoREST := &Context{ChannelID: types.Snowflake(200)}
	actNoREST := cNoREST.MessageSuite(types.Snowflake(500))

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

	actZeroMsg := c.MessageSuite(0)
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

	actZeroCh := c.MessageSuite(types.Snowflake(500))
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

	actValid := c.MessageSuite(types.Snowflake(500))
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

	act := c.MessageSuite(types.Snowflake(500))

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

	act := c.MessageSuite(types.Snowflake(500))

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

	act := c.MessageSuite(types.Snowflake(500))

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

func TestManageMessage_OptionsAndShortcuts(t *testing.T) {
	var pinned, deleted bool
	c, cleanup := setupMessageTestContext(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/channels/200/pins/500":
			pinned = true
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/channels/200/messages/500":
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer cleanup()

	err := c.ManageMessage("pin", types.Snowflake(500))
	if err != nil {
		t.Fatalf("unexpected pin shortcut error: %v", err)
	}
	if !pinned {
		t.Fatalf("expected pinned true")
	}

	err = c.ManageMessage(Options{
		Type:    "delete",
		Message: types.Snowflake(500),
	})
	if err != nil {
		t.Fatalf("unexpected delete options error: %v", err)
	}
	if !deleted {
		t.Fatalf("expected deleted true")
	}
}

func TestManageMessage_AutoResolveTargetMessage(t *testing.T) {
	var pinned, deleted bool
	c, cleanup := setupMessageTestContext(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && r.URL.Path == "/channels/200/pins/777":
			pinned = true
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/channels/200/messages/777":
			deleted = true
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer cleanup()

	targetID := types.Snowflake(777)
	c.Interaction = &types.Interaction{
		Data: &types.InteractionData{
			TargetID: &targetID,
		},
	}

	if err := c.ManageMessage("pin"); err != nil {
		t.Fatalf("unexpected error pinning context target: %v", err)
	}
	if !pinned {
		t.Fatal("expected target message 777 to be pinned")
	}

	if err := c.ManageMessage(Options{Type: "delete"}); err != nil {
		t.Fatalf("unexpected error deleting context target: %v", err)
	}
	if !deleted {
		t.Fatal("expected target message 777 to be deleted")
	}
}

func TestMessage_ReactionsAndEditing(t *testing.T) {
	var reacted []string
	var unreacted []string
	var clearedAll, clearedEmoji bool
	var editedContent string

	c, cleanup := setupMessageTestContext(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && strings.Contains(r.URL.Path, "/reactions/"):
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) >= 7 {
				reacted = append(reacted, parts[6])
			}
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && strings.HasSuffix(r.URL.Path, "/@me"):
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) >= 7 {
				unreacted = append(unreacted, parts[6])
			}
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && r.URL.Path == "/channels/200/messages/500/reactions":
			clearedAll = true
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete && strings.Contains(r.URL.Path, "/reactions/"):
			clearedEmoji = true
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPatch && r.URL.Path == "/channels/200/messages/500":
			var p rest.MessagePayload
			_ = json.NewDecoder(r.Body).Decode(&p)
			editedContent = p.Content
			_ = json.NewEncoder(w).Encode(types.Message{ID: 500, Content: p.Content})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer cleanup()

	err := c.Message(Options{
		Target: 500,
		Type:   "react",
		Emoji:  []string{"star", "<:pepe_cool:12345>", "<a:pepe_dance:67890>"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(reacted) != 3 || reacted[0] != "star" || reacted[1] != "pepe_cool:12345" || reacted[2] != "pepe_dance:67890" {
		t.Fatalf("unexpected reactions: %v", reacted)
	}

	err = c.Message(500, "unreact", "star")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(unreacted) != 1 || unreacted[0] != "star" {
		t.Fatalf("unexpected unreact: %v", unreacted)
	}

	err = c.Message(500, "clearreactions")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !clearedAll {
		t.Fatalf("expected clear all reactions")
	}

	err = c.Message(500, "clearreactions", "star")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !clearedEmoji {
		t.Fatalf("expected clear specific emoji")
	}

	err = c.Message(Options{
		Target:  500,
		Type:    "edit",
		Content: "edited message text",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if editedContent != "edited message text" {
		t.Fatalf("expected edited message text, got %s", editedContent)
	}

	err = c.Message(500, "edit", "positional edit text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if editedContent != "positional edit text" {
		t.Fatalf("expected positional edit text, got %s", editedContent)
	}
}

func TestMessagePurge_FilteringAndSingle(t *testing.T) {
	now := time.Now()
	oldTime := now.Add(-15 * 24 * time.Hour)
	oldID := types.Snowflake((oldTime.UnixMilli() - types.DiscordEpoch) << 22)

	recentTime := now.Add(-1 * time.Hour)
	recentID1 := types.Snowflake((recentTime.UnixMilli() - types.DiscordEpoch) << 22)
	recentID2 := recentID1 + 1

	var bulkDeleted []types.Snowflake
	var singleDeleted types.Snowflake

	c, cleanup := setupMessageTestContext(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "GET" && r.URL.Path == "/channels/200/messages":
			if r.URL.Query().Get("limit") == "1" {
				_ = json.NewEncoder(w).Encode([]types.Message{
					{ID: recentID1},
				})
				return
			}
			_ = json.NewEncoder(w).Encode([]types.Message{
				{ID: oldID},
				{ID: recentID1},
				{ID: recentID2},
			})
		case r.Method == "POST" && r.URL.Path == "/channels/200/messages/bulk-delete":
			var payload struct {
				Messages []types.Snowflake `json:"messages"`
			}
			_ = json.NewDecoder(r.Body).Decode(&payload)
			bulkDeleted = payload.Messages
			w.WriteHeader(http.StatusNoContent)
		case r.Method == "DELETE" && strings.HasPrefix(r.URL.Path, "/channels/200/messages/"):
			parts := strings.Split(r.URL.Path, "/")
			if len(parts) == 5 {
				id, _ := strconv.ParseUint(parts[4], 10, 64)
				singleDeleted = types.Snowflake(id)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	defer cleanup()

	err := c.Message("purge", 0)
	if !errors.Is(err, ErrInvalidPurgeCount) {
		t.Fatalf("expected ErrInvalidPurgeCount for 0, got %v", err)
	}

	err = c.Message("purge", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(bulkDeleted) != 2 {
		t.Fatalf("expected 2 messages in bulk delete (excluding old message), got %d", len(bulkDeleted))
	}
	for _, id := range bulkDeleted {
		if id == oldID {
			t.Fatalf("expected old message %v to be filtered out", oldID)
		}
	}

	err = c.Message("purge", 1)
	if err != nil {
		t.Fatalf("unexpected error on single purge: %v", err)
	}
	if singleDeleted != recentID1 {
		t.Fatalf("expected single delete to be called for %v, got %v", recentID1, singleDeleted)
	}
}
