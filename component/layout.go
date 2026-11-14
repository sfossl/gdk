package component

import (
	"github.com/sfossl/gda/types"
)

type SeparatorOpt func(*types.Component)
type ThumbnailOpt func(*types.Component)
type FileCardOpt func(*types.Component)
type MediaItemOpt func(*types.MediaGalleryItem)

func Text(content string, id ...int) types.Component {
	c := types.Component{
		Type:    types.ComponentTypeTextDisplay,
		Content: content,
	}
	if len(id) > 0 {
		val := id[0]
		c.ID = &val
	}
	return c
}

func Divider() types.Component {
	divider := true
	spacing := int(types.SeparatorSpacingSmall)
	return types.Component{
		Type:    types.ComponentTypeSeparator,
		Divider: &divider,
		Spacing: &spacing,
	}
}

func WithSpacing(spacing types.SeparatorSpacing) SeparatorOpt {
	return func(c *types.Component) {
		s := int(spacing)
		c.Spacing = &s
	}
}

func WithDivider(divider bool) SeparatorOpt {
	return func(c *types.Component) {
		c.Divider = &divider
	}
}

func WithSeparatorID(id int) SeparatorOpt {
	return func(c *types.Component) {
		c.ID = &id
	}
}

func Separator(opts ...SeparatorOpt) types.Component {
	divider := true
	spacing := int(types.SeparatorSpacingSmall)
	c := types.Component{
		Type:    types.ComponentTypeSeparator,
		Divider: &divider,
		Spacing: &spacing,
	}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

func WithThumbnailDescription(desc string) ThumbnailOpt {
	return func(c *types.Component) {
		c.Description = desc
	}
}

func WithThumbnailSpoiler(spoiler bool) ThumbnailOpt {
	return func(c *types.Component) {
		c.Spoiler = &spoiler
	}
}

func WithThumbnailID(id int) ThumbnailOpt {
	return func(c *types.Component) {
		c.ID = &id
	}
}

func Thumbnail(url string, opts ...ThumbnailOpt) types.Component {
	c := types.Component{
		Type:  types.ComponentTypeThumbnail,
		Media: &types.UnfurledMedia{URL: url},
	}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

func WithFileName(name string) FileCardOpt {
	return func(c *types.Component) {
		c.Name = name
	}
}

func WithFileSize(size int) FileCardOpt {
	return func(c *types.Component) {
		c.Size = &size
	}
}

func WithFileSpoiler(spoiler bool) FileCardOpt {
	return func(c *types.Component) {
		c.Spoiler = &spoiler
	}
}

func WithFileID(id int) FileCardOpt {
	return func(c *types.Component) {
		c.ID = &id
	}
}

func FileCard(url string, opts ...FileCardOpt) types.Component {
	c := types.Component{
		Type: types.ComponentTypeFile,
		File: &types.UnfurledMedia{URL: url},
	}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

func WithItemDescription(desc string) MediaItemOpt {
	return func(item *types.MediaGalleryItem) {
		item.Description = desc
	}
}

func WithItemSpoiler(spoiler bool) MediaItemOpt {
	return func(item *types.MediaGalleryItem) {
		item.Spoiler = spoiler
	}
}

func MediaItem(url string, opts ...MediaItemOpt) types.MediaGalleryItem {
	item := types.MediaGalleryItem{
		Media: types.UnfurledMedia{URL: url},
	}
	for _, opt := range opts {
		opt(&item)
	}
	return item
}

func MediaGallery(items ...types.MediaGalleryItem) types.Component {
	return types.Component{
		Type:  types.ComponentTypeMediaGallery,
		Items: items,
	}
}

type SectionBuilder struct {
	comp types.Component
}

func Section(texts ...string) *SectionBuilder {
	var children []types.Component
	for _, t := range texts {
		children = append(children, Text(t))
	}
	return &SectionBuilder{
		comp: types.Component{
			Type:       types.ComponentTypeSection,
			Components: children,
		},
	}
}

func (s *SectionBuilder) WithText(text string) *SectionBuilder {
	s.comp.Components = append(s.comp.Components, Text(text))
	return s
}

func (s *SectionBuilder) WithAccessory(accessory types.Component) *SectionBuilder {
	s.comp.Accessory = &accessory
	return s
}

func (s *SectionBuilder) WithID(id int) *SectionBuilder {
	s.comp.ID = &id
	return s
}

func (s *SectionBuilder) Build() types.Component {
	return s.comp
}

func SectionWithAccessory(accessory types.Component, texts ...string) types.Component {
	return Section(texts...).WithAccessory(accessory).Build()
}

type ContainerBuilder struct {
	comp types.Component
}

func Container(components ...types.Component) *ContainerBuilder {
	return &ContainerBuilder{
		comp: types.Component{
			Type:       types.ComponentTypeContainer,
			Components: components,
		},
	}
}

func (c *ContainerBuilder) WithID(id int) *ContainerBuilder {
	c.comp.ID = &id
	return c
}

func (c *ContainerBuilder) WithAccentColor(color any) *ContainerBuilder {
	if col, err := ResolveColor(color); err == nil {
		c.comp.AccentColor = &col
	}
	return c
}

func (c *ContainerBuilder) WithSpoiler(spoiler bool) *ContainerBuilder {
	c.comp.Spoiler = &spoiler
	return c
}

func (c *ContainerBuilder) Add(components ...types.Component) *ContainerBuilder {
	c.comp.Components = append(c.comp.Components, components...)
	return c
}

func (c *ContainerBuilder) Build() types.Component {
	return c.comp
}

func ContainerCard(color any, components ...types.Component) types.Component {
	return Container(components...).WithAccentColor(color).Build()
}
