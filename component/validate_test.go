package component

import (
	"strings"
	"testing"

	"github.com/sfossl/gda/types"
)

func TestValidateComponentsValidV1(t *testing.T) {
	btn := Button("click", "btn_1")
	rows := []types.Component{
		ActionRow(btn),
	}
	if err := ValidateComponents(rows); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestValidateComponentsValidLayout(t *testing.T) {
	btn := Button("click", "btn_1")
	card := Container(
		Section("Section 1").WithAccessory(btn).Build(),
		Divider(),
		Text("Some text"),
		FileCard("attachment://doc.pdf"),
		MediaGallery(MediaItem("https://example.com/1.png")),
		ActionRow(btn),
	).Build()

	if err := ValidateComponents([]types.Component{card}); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestValidateComponentsNestedContainers(t *testing.T) {
	inner := Container(Text("Inner")).Build()
	outer := Container(inner).Build()

	err := ValidateComponents([]types.Component{outer})
	if err != ErrNestedContainers {
		t.Fatalf("expected ErrNestedContainers, got %v", err)
	}
}

func TestValidateComponentsInvalidContainerChild(t *testing.T) {
	rawButton := Button("raw", "btn_raw")
	card := Container(rawButton).Build()

	err := ValidateComponents([]types.Component{card})
	if err != ErrButtonMustBeInRowOrAccessory {
		t.Fatalf("expected ErrButtonMustBeInRowOrAccessory, got %v", err)
	}

	selectMenu := StringSelect("sel", "choose", Option("1", "1"))
	cardWithSelect := Container(selectMenu).Build()
	err = ValidateComponents([]types.Component{cardWithSelect})
	if err != ErrInvalidContainerChild {
		t.Fatalf("expected ErrInvalidContainerChild, got %v", err)
	}
}

func TestValidateComponentsMaxCount(t *testing.T) {
	var children []types.Component
	for i := 0; i < 41; i++ {
		children = append(children, Text("item"))
	}
	card := Container(children...).Build()

	err := ValidateComponents([]types.Component{card})
	if err != ErrTooManyComponents {
		t.Fatalf("expected ErrTooManyComponents, got %v", err)
	}
}

func TestValidateComponentsMaxTextLength(t *testing.T) {
	t1 := Text(strings.Repeat("a", 2500))
	t2 := Text(strings.Repeat("b", 1600))
	card := Container(t1, t2).Build()

	err := ValidateComponents([]types.Component{card})
	if err != ErrTotalTextLengthExceeded {
		t.Fatalf("expected ErrTotalTextLengthExceeded, got %v", err)
	}
}

func TestValidateComponentsSectionLimits(t *testing.T) {
	sec := Section("t1", "t2", "t3", "t4").Build()
	card := Container(sec).Build()

	err := ValidateComponents([]types.Component{card})
	if err != ErrSectionTooManyTexts {
		t.Fatalf("expected ErrSectionTooManyTexts, got %v", err)
	}

	invalidAccessory := Text("not an accessory")
	secWithBadAccessory := Section("header").WithAccessory(invalidAccessory).Build()
	cardBad := Container(secWithBadAccessory).Build()

	err = ValidateComponents([]types.Component{cardBad})
	if err != ErrInvalidSectionAccessory {
		t.Fatalf("expected ErrInvalidSectionAccessory, got %v", err)
	}
}

func TestValidateComponentsMediaGalleryLimits(t *testing.T) {
	var items []types.MediaGalleryItem
	for i := 0; i < 11; i++ {
		items = append(items, MediaItem("https://example.com/img.png"))
	}
	gallery := MediaGallery(items...)
	card := Container(gallery).Build()

	err := ValidateComponents([]types.Component{card})
	if err != ErrMediaGalleryTooManyItems {
		t.Fatalf("expected ErrMediaGalleryTooManyItems, got %v", err)
	}
}

func TestValidateComponentsSeparatorSpacing(t *testing.T) {
	invalidSpacing := 99
	sep := types.Component{
		Type:    types.ComponentTypeSeparator,
		Spacing: &invalidSpacing,
	}
	card := Container(sep).Build()

	err := ValidateComponents([]types.Component{card})
	if err != ErrInvalidSeparatorSpacing {
		t.Fatalf("expected ErrInvalidSeparatorSpacing, got %v", err)
	}
}

func TestValidateComponentsPlacementGuards(t *testing.T) {
	thumb := Thumbnail("https://example.com/logo.png")
	if err := ValidateComponents([]types.Component{thumb}); err != ErrThumbnailMustBeAccessory {
		t.Fatalf("expected ErrThumbnailMustBeAccessory, got %v", err)
	}

	cardWithThumb := Container(thumb).Build()
	if err := ValidateComponents([]types.Component{cardWithThumb}); err != ErrThumbnailMustBeAccessory {
		t.Fatalf("expected ErrThumbnailMustBeAccessory, got %v", err)
	}

	btn := Button("click", "btn_1")
	containerAndLooseButton := []types.Component{Container(Text("hello")).Build(), btn}
	if err := ValidateComponents(containerAndLooseButton); err != ErrButtonMustBeInRowOrAccessory {
		t.Fatalf("expected ErrButtonMustBeInRowOrAccessory, got %v", err)
	}

	secBadChild := types.Component{
		Type:       types.ComponentTypeSection,
		Components: []types.Component{Divider()},
	}
	cardBadSec := Container(secBadChild).Build()
	if err := ValidateComponents([]types.Component{cardBadSec}); err != ErrSectionInvalidChild {
		t.Fatalf("expected ErrSectionInvalidChild, got %v", err)
	}
}
