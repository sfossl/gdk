package component

import (
	"errors"
	"fmt"
	"testing"

	"github.com/sfossl/gda/types"
)

func TestActionRowConstructors(t *testing.T) {
	b1 := Primary("b1", "Button 1")
	b2 := Secondary("b2", "Button 2")

	row := ActionRow(b1, b2)
	if row.Type != types.ComponentTypeActionRow || len(row.Components) != 2 {
		t.Fatalf("unexpected action row %v", row)
	}

	rowAlias := Row(b1, b2)
	if rowAlias.Type != types.ComponentTypeActionRow || len(rowAlias.Components) != 2 {
		t.Fatalf("unexpected row alias %v", rowAlias)
	}
}

func TestPackRows_ButtonsOnly(t *testing.T) {
	var btns []types.Component
	for i := 0; i < 7; i++ {
		btns = append(btns, Primary(fmt.Sprintf("b_%d", i), fmt.Sprintf("B%d", i)))
	}

	rows := PackRows(btns...)
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if len(rows[0].Components) != 5 {
		t.Fatalf("expected 5 buttons in row 0, got %d", len(rows[0].Components))
	}
	if len(rows[1].Components) != 2 {
		t.Fatalf("expected 2 buttons in row 1, got %d", len(rows[1].Components))
	}

	if err := ValidateRows(rows); err != nil {
		t.Fatalf("expected valid rows, got error %v", err)
	}
}

func TestPackRows_FailSoftTruncation(t *testing.T) {
	var btns []types.Component
	for i := 0; i < 30; i++ {
		btns = append(btns, Primary(fmt.Sprintf("b_%d", i), fmt.Sprintf("B%d", i)))
	}

	rows := PackRows(btns...)
	if len(rows) != 5 {
		t.Fatalf("expected capped at 5 rows, got %d", len(rows))
	}

	totalButtons := 0
	for _, r := range rows {
		totalButtons += len(r.Components)
	}
	if totalButtons != 25 {
		t.Fatalf("expected exactly 25 buttons packed, got %d", totalButtons)
	}

	if err := ValidateRows(rows); err != nil {
		t.Fatalf("expected valid rows, got error %v", err)
	}
}

func TestPackRows_MixedButtonsAndSelects(t *testing.T) {
	b1 := Primary("b1", "B1")
	b2 := Primary("b2", "B2")
	sel := StringSelect("sel_1", "Choose", Option("O1", "v1"))
	b3 := Primary("b3", "B3")
	b4 := Primary("b4", "B4")

	rows := PackRows(b1, b2, sel, b3, b4)
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}

	if len(rows[0].Components) != 2 || rows[0].Components[0].CustomID != "b1" {
		t.Fatalf("unexpected row 0 %v", rows[0])
	}
	if len(rows[1].Components) != 1 || rows[1].Components[0].CustomID != "sel_1" {
		t.Fatalf("unexpected row 1 %v", rows[1])
	}
	if len(rows[2].Components) != 2 || rows[2].Components[0].CustomID != "b3" {
		t.Fatalf("unexpected row 2 %v", rows[2])
	}

	if err := ValidateRows(rows); err != nil {
		t.Fatalf("expected valid rows, got error %v", err)
	}
}

func TestPackRows_PreExistingActionRow(t *testing.T) {
	b1 := Primary("b1", "B1")
	preRow := ActionRow(Secondary("b2", "B2"))
	b3 := Primary("b3", "B3")

	rows := PackRows(b1, preRow, b3)
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
}

func TestValidateRows_Errors(t *testing.T) {
	var rows6 []types.Component
	for i := 0; i < 6; i++ {
		rows6 = append(rows6, ActionRow(Primary(fmt.Sprintf("b%d", i), "B")))
	}
	if err := ValidateRows(rows6); !errors.Is(err, ErrTooManyActionRows) {
		t.Fatalf("expected ErrTooManyActionRows, got %v", err)
	}

	notRow := []types.Component{Primary("b1", "B")}
	if err := ValidateRows(notRow); !errors.Is(err, ErrNonActionRowAtRoot) {
		t.Fatalf("expected ErrNonActionRowAtRoot, got %v", err)
	}

	emptyRow := []types.Component{ActionRow()}
	if err := ValidateRows(emptyRow); !errors.Is(err, ErrEmptyActionRow) {
		t.Fatalf("expected ErrEmptyActionRow, got %v", err)
	}

	var btns6 []types.Component
	for i := 0; i < 6; i++ {
		btns6 = append(btns6, Primary(fmt.Sprintf("b%d", i), "B"))
	}
	tooManyBtns := []types.Component{ActionRow(btns6...)}
	if err := ValidateRows(tooManyBtns); !errors.Is(err, ErrRowTooManyButtons) {
		t.Fatalf("expected ErrRowTooManyButtons, got %v", err)
	}

	mixed := []types.Component{ActionRow(Primary("b1", "B"), StringSelect("s1", "S"))}
	if err := ValidateRows(mixed); !errors.Is(err, ErrRowMixedComponents) {
		t.Fatalf("expected ErrRowMixedComponents, got %v", err)
	}

	multiSelect := []types.Component{ActionRow(StringSelect("s1", "S1"), StringSelect("s2", "S2"))}
	if err := ValidateRows(multiSelect); !errors.Is(err, ErrRowMultipleSelects) {
		t.Fatalf("expected ErrRowMultipleSelects, got %v", err)
	}

	multiInput := []types.Component{ActionRow(TextInputShort("i1", "L1"), TextInputShort("i2", "L2"))}
	if err := ValidateRows(multiInput); !errors.Is(err, ErrRowMultipleInputs) {
		t.Fatalf("expected ErrRowMultipleInputs, got %v", err)
	}
}

