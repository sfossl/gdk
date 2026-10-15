package component

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/sfossl/gda/types"
)

func TestStress_EmbedExactRuneBoundaries(t *testing.T) {
	exactTitle := strings.Repeat("T", 256)
	bTitlePass := NewEmbed().SetTitle(exactTitle)
	if _, err := bTitlePass.Build(); err != nil {
		t.Fatalf("expected 256 rune title to pass, got %v", err)
	}

	overTitle := strings.Repeat("T", 257)
	bTitleFail := NewEmbed().SetTitle(overTitle)
	if _, err := bTitleFail.Build(); !errors.Is(err, ErrEmbedTitleTooLong) {
		t.Fatalf("expected ErrEmbedTitleTooLong, got %v", err)
	}

	exactDesc := strings.Repeat("D", 4096)
	bDescPass := NewEmbed().SetDescription(exactDesc)
	if _, err := bDescPass.Build(); err != nil {
		t.Fatalf("expected 4096 rune desc to pass, got %v", err)
	}

	overDesc := strings.Repeat("D", 4097)
	bDescFail := NewEmbed().SetDescription(overDesc)
	if _, err := bDescFail.Build(); !errors.Is(err, ErrEmbedDescriptionTooLong) {
		t.Fatalf("expected ErrEmbedDescriptionTooLong, got %v", err)
	}

	exactFieldName := strings.Repeat("N", 256)
	exactFieldValue := strings.Repeat("V", 1024)
	bFieldPass := NewEmbed().AddField(exactFieldName, exactFieldValue)
	if _, err := bFieldPass.Build(); err != nil {
		t.Fatalf("expected 256/1024 rune field to pass, got %v", err)
	}

	overFieldName := strings.Repeat("N", 257)
	bFieldNameFail := NewEmbed().AddField(overFieldName, exactFieldValue)
	if _, err := bFieldNameFail.Build(); !errors.Is(err, ErrEmbedFieldNameTooLong) {
		t.Fatalf("expected ErrEmbedFieldNameTooLong, got %v", err)
	}

	overFieldValue := strings.Repeat("V", 1025)
	bFieldValueFail := NewEmbed().AddField(exactFieldName, overFieldValue)
	if _, err := bFieldValueFail.Build(); !errors.Is(err, ErrEmbedFieldValueTooLong) {
		t.Fatalf("expected ErrEmbedFieldValueTooLong, got %v", err)
	}
}

func TestStress_EmbedTotalLengthBoundary(t *testing.T) {
	b6000 := NewEmbed().
		SetTitle(strings.Repeat("a", 256)).
		SetDescription(strings.Repeat("b", 4096)).
		SetFooter(strings.Repeat("c", 1392)).
		SetAuthor(strings.Repeat("d", 256))

	embed6000, err := b6000.Build()
	if err != nil {
		t.Fatalf("expected exactly 6000 characters to pass, got %v", err)
	}
	if embed6000 == nil {
		t.Fatal("expected non-nil embed")
	}

	b6001 := NewEmbed().
		SetTitle(strings.Repeat("a", 256)).
		SetDescription(strings.Repeat("b", 4096)).
		SetFooter(strings.Repeat("c", 1393)).
		SetAuthor(strings.Repeat("d", 256))

	if _, err := b6001.Build(); !errors.Is(err, ErrEmbedTotalLengthTooLong) {
		t.Fatalf("expected ErrEmbedTotalLengthTooLong for 6001 characters, got %v", err)
	}
}

func TestStress_EmbedFieldCountLimit(t *testing.T) {
	b25 := NewEmbed()
	for i := 0; i < 25; i++ {
		b25.AddField(fmt.Sprintf("f%d", i), "val")
	}
	if _, err := b25.Build(); err != nil {
		t.Fatalf("expected 25 fields to pass, got %v", err)
	}

	b26 := NewEmbed()
	for i := 0; i < 26; i++ {
		b26.AddField(fmt.Sprintf("f%d", i), "val")
	}
	if _, err := b26.Build(); !errors.Is(err, ErrEmbedTooManyFields) {
		t.Fatalf("expected ErrEmbedTooManyFields for 26 fields, got %v", err)
	}
}

func TestStress_PackRows_25Buttons(t *testing.T) {
	var buttons []types.Component
	for i := 0; i < 30; i++ {
		buttons = append(buttons, Primary(fmt.Sprintf("btn_%d", i), fmt.Sprintf("Btn %d", i)))
	}

	rows := PackRows(buttons...)
	if len(rows) != 5 {
		t.Fatalf("expected exactly 5 action rows, got %d", len(rows))
	}

	totalButtons := 0
	for _, row := range rows {
		if len(row.Components) != 5 {
			t.Fatalf("expected 5 buttons per row, got %d", len(row.Components))
		}
		totalButtons += len(row.Components)
	}

	if totalButtons != 25 {
		t.Fatalf("expected 25 packed buttons, got %d", totalButtons)
	}

	if err := ValidateRows(rows); err != nil {
		t.Fatalf("expected 5x5 packed buttons to pass validation, got %v", err)
	}
}

func TestStress_SelectOptionBounds(t *testing.T) {
	emptySelectRow := []types.Component{
		ActionRow(types.Component{
			Type:     types.ComponentTypeStringSelect,
			CustomID: "sel",
			Options:  nil,
		}),
	}
	if err := ValidateRows(emptySelectRow); !errors.Is(err, ErrSelectMissingOptions) {
		t.Fatalf("expected ErrSelectMissingOptions for empty options, got %v", err)
	}

	var opts25 []types.SelectOption
	for i := 0; i < 25; i++ {
		opts25 = append(opts25, SelectOption(fmt.Sprintf("L%d", i), fmt.Sprintf("V%d", i)))
	}
	passSelectRow := []types.Component{
		ActionRow(StringSelect("sel", "placeholder", opts25...)),
	}
	if err := ValidateRows(passSelectRow); err != nil {
		t.Fatalf("expected 25 select options to pass, got %v", err)
	}

	opts26 := append(opts25, SelectOption("L26", "V26"))
	failSelectRow := []types.Component{
		ActionRow(StringSelect("sel", "placeholder", opts26...)),
	}
	if err := ValidateRows(failSelectRow); !errors.Is(err, ErrSelectTooManyOptions) {
		t.Fatalf("expected ErrSelectTooManyOptions for 26 options, got %v", err)
	}
}

func TestStress_ConcurrentEmbedAndColorResolutions(t *testing.T) {
	colors := []any{
		"blurple",
		"#5865F2",
		"0x57F287",
		0xEB459E,
		"fuchsia",
		"dark",
		"#fff",
		"0XED4245",
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for _, col := range colors {
				resolved, err := ResolveColor(col)
				if err != nil {
					t.Errorf("worker %d unexpected error resolving %v: %v", workerID, col, err)
				}
				if resolved == 0 && col != "black" {
					t.Errorf("worker %d unexpected 0 color for %v", workerID, col)
				}

				b := NewEmbed().
					SetTitle("Concurrent Title").
					SetDescription("Concurrent Description").
					SetColor(col).
					AddField("Worker", fmt.Sprintf("%d", workerID))

				embed, err := b.Build()
				if err != nil || embed == nil {
					t.Errorf("worker %d failed to build embed: %v", workerID, err)
				}
			}
		}(i)
	}
	wg.Wait()
}

