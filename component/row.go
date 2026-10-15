package component

import (
	"errors"

	"github.com/sfossl/gda/types"
)

const (
	MaxActionRows       = 5
	MaxButtonsPerRow    = 5
	MaxSelectsPerRow    = 1
	MaxTextInputsPerRow = 1
)

var (
	ErrTooManyActionRows  = errors.New("components exceed maximum 5 action rows")
	ErrRowTooManyButtons  = errors.New("action row exceeds maximum 5 buttons")
	ErrRowMixedComponents = errors.New("action row cannot mix buttons and select menus")
	ErrRowMultipleSelects = errors.New("action row cannot contain multiple select menus")
	ErrRowMultipleInputs  = errors.New("action row cannot contain multiple text inputs")
	ErrEmptyActionRow     = errors.New("action row cannot be empty")
	ErrNonActionRowAtRoot = errors.New("top level message component must be an action row")
)

func ActionRow(components ...types.Component) types.Component {
	return types.Component{
		Type:       types.ComponentTypeActionRow,
		Components: components,
	}
}

func Row(components ...types.Component) types.Component {
	return ActionRow(components...)
}

func PackRows(components ...types.Component) []types.Component {
	var rows []types.Component
	var pendingButtons []types.Component

	flushButtons := func() {
		if len(pendingButtons) > 0 {
			if len(rows) < MaxActionRows {
				rows = append(rows, ActionRow(pendingButtons...))
			}
			pendingButtons = nil
		}
	}

	for _, c := range components {
		if len(rows) >= MaxActionRows && (isSelectMenu(c) || isTextInput(c) || c.Type == types.ComponentTypeActionRow) {
			continue
		}

		if c.Type == types.ComponentTypeActionRow {
			flushButtons()
			if len(rows) < MaxActionRows {
				rows = append(rows, c)
			}
			continue
		}

		if isButton(c) {
			if len(pendingButtons) == MaxButtonsPerRow {
				flushButtons()
			}
			if len(rows) < MaxActionRows {
				pendingButtons = append(pendingButtons, c)
			}
			continue
		}

		if isSelectMenu(c) || isTextInput(c) {
			flushButtons()
			if len(rows) < MaxActionRows {
				rows = append(rows, ActionRow(c))
			}
			continue
		}
	}

	flushButtons()
	return rows
}

func ValidateRows(rows []types.Component) error {
	if len(rows) > MaxActionRows {
		return ErrTooManyActionRows
	}

	for _, row := range rows {
		if row.Type != types.ComponentTypeActionRow {
			return ErrNonActionRowAtRoot
		}

		if len(row.Components) == 0 {
			return ErrEmptyActionRow
		}

		buttons := 0
		selects := 0
		inputs := 0

		for _, item := range row.Components {
			switch {
			case isButton(item):
				buttons++
			case isSelectMenu(item):
				selects++
			case isTextInput(item):
				inputs++
			}
		}

		if buttons > 0 && (selects > 0 || inputs > 0) {
			return ErrRowMixedComponents
		}
		if buttons > MaxButtonsPerRow {
			return ErrRowTooManyButtons
		}
		if selects > MaxSelectsPerRow {
			return ErrRowMultipleSelects
		}
		if inputs > MaxTextInputsPerRow {
			return ErrRowMultipleInputs
		}
	}

	return nil
}

func isButton(c types.Component) bool {
	return c.Type == types.ComponentTypeButton
}

func isSelectMenu(c types.Component) bool {
	switch c.Type {
	case types.ComponentTypeStringSelect,
		types.ComponentTypeUserSelect,
		types.ComponentTypeRoleSelect,
		types.ComponentTypeMentionableSelect,
		types.ComponentTypeChannelSelect:
		return true
	default:
		return false
	}
}

func isTextInput(c types.Component) bool {
	return c.Type == types.ComponentTypeTextInput
}

