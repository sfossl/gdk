package component

import (
	"testing"

	"github.com/sfossl/gda/types"
)

func TestText(t *testing.T) {
	c := Text("hello world")
	if c.Type != types.ComponentTypeTextDisplay {
		t.Fatalf("expected text display type, got %d", c.Type)
	}
	if c.Content != "hello world" {
		t.Fatalf("expected content hello world, got %s", c.Content)
	}
}

func TestDividerAndSeparator(t *testing.T) {
	d := Divider()
	if d.Type != types.ComponentTypeSeparator {
		t.Fatalf("expected separator type")
	}
	if d.Divider == nil || !*d.Divider {
		t.Fatalf("expected divider true")
	}
	if d.Spacing == nil || *d.Spacing != int(types.SeparatorSpacingSmall) {
		t.Fatalf("expected small spacing")
	}

	hasDiv := false
	sep := Separator(WithDivider(hasDiv), WithSpacing(types.SeparatorSpacingLarge))
	if sep.Divider == nil || *sep.Divider {
		t.Fatalf("expected divider false")
	}
	if sep.Spacing == nil || *sep.Spacing != int(types.SeparatorSpacingLarge) {
		t.Fatalf("expected large spacing")
	}
}

func TestThumbnail(t *testing.T) {
	thumb := Thumbnail("https://example.com/logo.png",
		WithThumbnailDescription("logo"),
		WithThumbnailSpoiler(true),
	)
	if thumb.Type != types.ComponentTypeThumbnail {
		t.Fatalf("expected thumbnail type")
	}
	if thumb.Media == nil || thumb.Media.URL != "https://example.com/logo.png" {
		t.Fatalf("expected url")
	}
	if thumb.Description != "logo" {
		t.Fatalf("expected description logo")
	}
	if thumb.Spoiler == nil || !*thumb.Spoiler {
		t.Fatalf("expected spoiler true")
	}
}

func TestFileCard(t *testing.T) {
	fc := FileCard("attachment://notes.pdf",
		WithFileName("notes.pdf"),
		WithFileSize(1024),
		WithFileSpoiler(true),
	)
	if fc.Type != types.ComponentTypeFile {
		t.Fatalf("expected file type")
	}
	if fc.File == nil || fc.File.URL != "attachment://notes.pdf" {
		t.Fatalf("expected file url")
	}
	if fc.Name != "notes.pdf" {
		t.Fatalf("expected name notes.pdf")
	}
	if fc.Size == nil || *fc.Size != 1024 {
		t.Fatalf("expected size 1024")
	}
	if fc.Spoiler == nil || !*fc.Spoiler {
		t.Fatalf("expected spoiler true")
	}
}

func TestMediaGallery(t *testing.T) {
	item1 := MediaItem("https://example.com/1.png", WithItemDescription("one"), WithItemSpoiler(false))
	item2 := MediaItem("https://example.com/2.png", WithItemSpoiler(true))
	gallery := MediaGallery(item1, item2)

	if gallery.Type != types.ComponentTypeMediaGallery {
		t.Fatalf("expected gallery type")
	}
	if len(gallery.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(gallery.Items))
	}
	if gallery.Items[0].Description != "one" || gallery.Items[1].Spoiler != true {
		t.Fatalf("item values mismatch")
	}
}

func TestSection(t *testing.T) {
	btn := Button("click", "btn_1")
	sec := Section("First text").
		WithText("Second text").
		WithAccessory(btn).
		Build()

	if sec.Type != types.ComponentTypeSection {
		t.Fatalf("expected section type")
	}
	if len(sec.Components) != 2 {
		t.Fatalf("expected 2 text children, got %d", len(sec.Components))
	}
	if sec.Accessory == nil || sec.Accessory.CustomID != "btn_1" {
		t.Fatalf("expected button accessory")
	}

	secShort := SectionWithAccessory(btn, "Header", "Subtitle")
	if secShort.Type != types.ComponentTypeSection || len(secShort.Components) != 2 || secShort.Accessory == nil {
		t.Fatalf("expected section with accessory and 2 texts")
	}
}

func TestContainer(t *testing.T) {
	card := Container(
		Text("Inner text"),
		Divider(),
	).
		WithAccentColor("blurple").
		WithSpoiler(true).
		Add(Text("Added text")).
		Build()

	if card.Type != types.ComponentTypeContainer {
		t.Fatalf("expected container type")
	}
	if card.AccentColor == nil || *card.AccentColor != ColorBlurple {
		t.Fatalf("expected blurple accent color, got %v", card.AccentColor)
	}
	if card.Spoiler == nil || !*card.Spoiler {
		t.Fatalf("expected spoiler true")
	}
	if len(card.Components) != 3 {
		t.Fatalf("expected 3 children, got %d", len(card.Components))
	}

	quickCard := ContainerCard("#FF0000", Text("Red card"))
	if quickCard.Type != types.ComponentTypeContainer || quickCard.AccentColor == nil || *quickCard.AccentColor != 0xFF0000 {
		t.Fatalf("expected red container card")
	}
}
