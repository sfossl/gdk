package component

import (
	"errors"
	"unicode/utf8"

	"github.com/sfossl/gda/types"
)

const (
	MaxTotalComponents     = 40
	MaxTotalTextLength     = 4000
	MaxSectionTextDisplays = 3
	MaxMediaGalleryItems   = 10
)

var (
	ErrTooManyComponents            = errors.New("total components exceed maximum 40")
	ErrTotalTextLengthExceeded      = errors.New("total text display content exceeds maximum 4000 characters")
	ErrNestedContainers             = errors.New("container component cannot be nested inside another container")
	ErrEmptyContainer               = errors.New("container component cannot be empty")
	ErrEmptySection                 = errors.New("section component must contain at least 1 text display")
	ErrEmptyMediaGallery            = errors.New("media gallery must contain at least 1 item")
	ErrInvalidContainerChild        = errors.New("invalid component type inside container")
	ErrSectionTooManyTexts          = errors.New("section exceeds maximum 3 text displays")
	ErrSectionInvalidChild          = errors.New("section can only contain text display components")
	ErrInvalidSectionAccessory      = errors.New("section accessory must be a button or thumbnail")
	ErrMediaGalleryTooManyItems     = errors.New("media gallery exceeds maximum 10 items")
	ErrInvalidSeparatorSpacing      = errors.New("separator spacing must be small (1) or large (2)")
	ErrThumbnailMustBeAccessory     = errors.New("thumbnail component can only be used as a section accessory")
	ErrButtonMustBeInRowOrAccessory = errors.New("button component must be in an action row or section accessory")
	ErrInvalidRootComponent         = errors.New("invalid component type at message root")
	ErrEmbedsNotAllowedWithLayouts  = errors.New("cannot mix legacy embeds with layout components")
	ErrContentNotAllowedWithLayouts = errors.New("cannot mix top-level content with layout components; use Text instead")
)

func IsLayoutComponent(t types.ComponentType) bool {
	switch t {
	case types.ComponentTypeSection,
		types.ComponentTypeTextDisplay,
		types.ComponentTypeThumbnail,
		types.ComponentTypeMediaGallery,
		types.ComponentTypeFile,
		types.ComponentTypeSeparator,
		types.ComponentTypeContainer:
		return true
	default:
		return false
	}
}

func HasLayoutComponents(components []types.Component) bool {
	for _, c := range components {
		if IsLayoutComponent(c.Type) {
			return true
		}
		if len(c.Components) > 0 && HasLayoutComponents(c.Components) {
			return true
		}
		if c.Accessory != nil && IsLayoutComponent(c.Accessory.Type) {
			return true
		}
	}
	return false
}

func ValidateComponents(components []types.Component) error {
	if len(components) == 0 {
		return nil
	}

	if !HasLayoutComponents(components) {
		return ValidateRows(components)
	}

	totalCount := 0
	totalTextRunes := 0

	var walk func(c types.Component, isRoot, isDirectContainerChild, insideContainer bool) error
	walk = func(c types.Component, isRoot, isDirectContainerChild, insideContainer bool) error {
		totalCount++
		if totalCount > MaxTotalComponents {
			return ErrTooManyComponents
		}

		if c.Type == types.ComponentTypeContainer {
			if insideContainer {
				return ErrNestedContainers
			}
			if len(c.Components) == 0 {
				return ErrEmptyContainer
			}
			for _, child := range c.Components {
				if err := walk(child, false, true, true); err != nil {
					return err
				}
			}
			return nil
		}

		if isRoot {
			switch c.Type {
			case types.ComponentTypeSection,
				types.ComponentTypeTextDisplay,
				types.ComponentTypeMediaGallery,
				types.ComponentTypeFile,
				types.ComponentTypeSeparator,
				types.ComponentTypeActionRow:
			case types.ComponentTypeThumbnail:
				return ErrThumbnailMustBeAccessory
			case types.ComponentTypeButton:
				return ErrButtonMustBeInRowOrAccessory
			default:
				return ErrInvalidRootComponent
			}
		}

		if isDirectContainerChild {
			switch c.Type {
			case types.ComponentTypeSection,
				types.ComponentTypeTextDisplay,
				types.ComponentTypeMediaGallery,
				types.ComponentTypeFile,
				types.ComponentTypeSeparator,
				types.ComponentTypeActionRow:
			case types.ComponentTypeThumbnail:
				return ErrThumbnailMustBeAccessory
			case types.ComponentTypeButton:
				return ErrButtonMustBeInRowOrAccessory
			default:
				return ErrInvalidContainerChild
			}
		}

		if c.Type == types.ComponentTypeTextDisplay {
			totalTextRunes += utf8.RuneCountInString(c.Content)
			if totalTextRunes > MaxTotalTextLength {
				return ErrTotalTextLengthExceeded
			}
		}

		if c.Type == types.ComponentTypeSection {
			if len(c.Components) == 0 {
				return ErrEmptySection
			}
			if len(c.Components) > MaxSectionTextDisplays {
				return ErrSectionTooManyTexts
			}
			for _, child := range c.Components {
				if child.Type != types.ComponentTypeTextDisplay {
					return ErrSectionInvalidChild
				}
				totalCount++
				if totalCount > MaxTotalComponents {
					return ErrTooManyComponents
				}
				totalTextRunes += utf8.RuneCountInString(child.Content)
				if totalTextRunes > MaxTotalTextLength {
					return ErrTotalTextLengthExceeded
				}
			}
			if c.Accessory != nil {
				totalCount++
				if totalCount > MaxTotalComponents {
					return ErrTooManyComponents
				}
				if c.Accessory.Type != types.ComponentTypeButton && c.Accessory.Type != types.ComponentTypeThumbnail {
					return ErrInvalidSectionAccessory
				}
			}
			return nil
		}

		if c.Type == types.ComponentTypeMediaGallery {
			if len(c.Items) == 0 {
				return ErrEmptyMediaGallery
			}
			if len(c.Items) > MaxMediaGalleryItems {
				return ErrMediaGalleryTooManyItems
			}
		}

		if c.Type == types.ComponentTypeSeparator && c.Spacing != nil {
			if *c.Spacing != int(types.SeparatorSpacingSmall) && *c.Spacing != int(types.SeparatorSpacingLarge) {
				return ErrInvalidSeparatorSpacing
			}
		}

		if c.Type == types.ComponentTypeActionRow {
			if err := ValidateRows([]types.Component{c}); err != nil {
				return err
			}
			return nil
		}

		for _, child := range c.Components {
			if err := walk(child, false, false, insideContainer); err != nil {
				return err
			}
		}

		return nil
	}

	for _, c := range components {
		if err := walk(c, true, false, false); err != nil {
			return err
		}
	}

	return nil
}
