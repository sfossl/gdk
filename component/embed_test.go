package component

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/sfossl/gda/types"
)

func TestEmbedBuilder_FluentSuccess(t *testing.T) {
	now := time.Now().UTC()
	b := NewEmbed().
		SetTitle("Test Title").
		SetDescription("Test Description").
		SetURL("https://example.com").
		SetColor("blurple").
		SetTimestamp(now).
		SetFooter("Footer Text", "https://example.com/footer.png").
		SetAuthor("Author Name", "https://example.com/icon.png", "https://example.com/author").
		SetImage("https://example.com/image.png").
		SetThumbnail("https://example.com/thumb.png").
		AddField("Field 1", "Value 1", true).
		AddFields(types.EmbedField{Name: "Field 2", Value: "Value 2", Inline: false})

	embed, err := b.Build()
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if embed.Title != "Test Title" {
		t.Fatalf("expected title 'Test Title', got %s", embed.Title)
	}
	if embed.Description != "Test Description" {
		t.Fatalf("expected description 'Test Description', got %s", embed.Description)
	}
	if embed.URL != "https://example.com" {
		t.Fatalf("expected URL, got %s", embed.URL)
	}
	if embed.Color != ColorBlurple {
		t.Fatalf("expected blurple color, got %d", embed.Color)
	}
	if embed.Timestamp == nil || *embed.Timestamp != now {
		t.Fatalf("expected timestamp %v, got %v", now, embed.Timestamp)
	}
	if embed.Footer == nil || embed.Footer.Text != "Footer Text" || embed.Footer.IconURL != "https://example.com/footer.png" {
		t.Fatalf("unexpected footer %v", embed.Footer)
	}
	if embed.Author == nil || embed.Author.Name != "Author Name" || embed.Author.URL != "https://example.com/author" {
		t.Fatalf("unexpected author %v", embed.Author)
	}
	if embed.Image == nil || embed.Image.URL != "https://example.com/image.png" {
		t.Fatalf("unexpected image %v", embed.Image)
	}
	if embed.Thumbnail == nil || embed.Thumbnail.URL != "https://example.com/thumb.png" {
		t.Fatalf("unexpected thumbnail %v", embed.Thumbnail)
	}
	if len(embed.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(embed.Fields))
	}
	if embed.Fields[0].Name != "Field 1" || !embed.Fields[0].Inline {
		t.Fatalf("unexpected field 0 %v", embed.Fields[0])
	}

	toEmbed := b.ToEmbed()
	if toEmbed.Title != "Test Title" {
		t.Fatalf("expected ToEmbed title 'Test Title', got %s", toEmbed.Title)
	}
}

func TestEmbedBuilder_ClearFields(t *testing.T) {
	b := NewEmbed().
		AddField("F1", "V1").
		AddField("F2", "V2")

	if len(b.embed.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(b.embed.Fields))
	}

	b.ClearFields()
	if len(b.embed.Fields) != 0 {
		t.Fatalf("expected 0 fields, got %d", len(b.embed.Fields))
	}
}

func TestEmbedBuilder_SetTimestamp(t *testing.T) {
	now := time.Now().UTC()
	b := NewEmbed()

	b.SetTimestamp(&now)
	if b.embed.Timestamp == nil || *b.embed.Timestamp != now {
		t.Fatalf("expected pointer timestamp to match")
	}

	b.SetTimestamp(nil)
	if b.embed.Timestamp != nil {
		t.Fatalf("expected nil timestamp")
	}

	b.SetTimestamp(types.Snowflake(175928847299117063))
	if b.embed.Timestamp == nil {
		t.Fatalf("expected snowflake timestamp to resolve")
	}

	b.SetTimestamp("invalid-timestamp")
	if _, err := b.Build(); err == nil {
		t.Fatalf("expected error from invalid timestamp string")
	}
}

func TestEmbedBuilder_SetColorError(t *testing.T) {
	b := NewEmbed().SetColor("invalid-color-name")
	if _, err := b.Build(); !errors.Is(err, ErrInvalidColor) {
		t.Fatalf("expected ErrInvalidColor, got %v", err)
	}
}

func TestEmbedBuilder_Limits(t *testing.T) {
	long257 := strings.Repeat("a", 257)
	long4097 := strings.Repeat("b", 4097)
	long1025 := strings.Repeat("c", 1025)
	long2049 := strings.Repeat("d", 2049)

	if _, err := NewEmbed().SetTitle(long257).Build(); !errors.Is(err, ErrEmbedTitleTooLong) {
		t.Fatalf("expected ErrEmbedTitleTooLong, got %v", err)
	}

	if _, err := NewEmbed().SetDescription(long4097).Build(); !errors.Is(err, ErrEmbedDescriptionTooLong) {
		t.Fatalf("expected ErrEmbedDescriptionTooLong, got %v", err)
	}

	bFields := NewEmbed()
	for i := 0; i < 26; i++ {
		bFields.AddField("name", "value")
	}
	if _, err := bFields.Build(); !errors.Is(err, ErrEmbedTooManyFields) {
		t.Fatalf("expected ErrEmbedTooManyFields, got %v", err)
	}

	if _, err := NewEmbed().AddField("", "value").Build(); !errors.Is(err, ErrEmbedFieldMissingName) {
		t.Fatalf("expected ErrEmbedFieldMissingName, got %v", err)
	}

	if _, err := NewEmbed().AddField(long257, "value").Build(); !errors.Is(err, ErrEmbedFieldNameTooLong) {
		t.Fatalf("expected ErrEmbedFieldNameTooLong, got %v", err)
	}

	if _, err := NewEmbed().AddField("name", "").Build(); !errors.Is(err, ErrEmbedFieldMissingValue) {
		t.Fatalf("expected ErrEmbedFieldMissingValue, got %v", err)
	}

	if _, err := NewEmbed().AddField("name", long1025).Build(); !errors.Is(err, ErrEmbedFieldValueTooLong) {
		t.Fatalf("expected ErrEmbedFieldValueTooLong, got %v", err)
	}

	if _, err := NewEmbed().SetFooter(long2049).Build(); !errors.Is(err, ErrEmbedFooterTooLong) {
		t.Fatalf("expected ErrEmbedFooterTooLong, got %v", err)
	}

	if _, err := NewEmbed().SetAuthor(long257).Build(); !errors.Is(err, ErrEmbedAuthorTooLong) {
		t.Fatalf("expected ErrEmbedAuthorTooLong, got %v", err)
	}

	bTotal := NewEmbed().
		SetDescription(strings.Repeat("x", 4000)).
		AddField("name", strings.Repeat("y", 1000)).
		AddField("name2", strings.Repeat("z", 1000)).
		AddField("name3", "extra-chars")
	if _, err := bTotal.Build(); !errors.Is(err, ErrEmbedTotalLengthTooLong) {
		t.Fatalf("expected ErrEmbedTotalLengthTooLong, got %v", err)
	}
}

