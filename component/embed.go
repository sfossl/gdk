package component

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sfossl/gda/types"
	"github.com/sfossl/gdk/util"
)

var (
	ErrEmbedTitleTooLong       = errors.New("embed title exceeds maximum 256 characters")
	ErrEmbedDescriptionTooLong = errors.New("embed description exceeds maximum 4096 characters")
	ErrEmbedTooManyFields      = errors.New("embed exceeds maximum 25 fields")
	ErrEmbedFieldNameTooLong   = errors.New("embed field name exceeds maximum 256 characters")
	ErrEmbedFieldValueTooLong  = errors.New("embed field value exceeds maximum 1024 characters")
	ErrEmbedFieldMissingName   = errors.New("embed field name cannot be empty")
	ErrEmbedFieldMissingValue  = errors.New("embed field value cannot be empty")
	ErrEmbedFooterTooLong      = errors.New("embed footer text exceeds maximum 2048 characters")
	ErrEmbedAuthorTooLong      = errors.New("embed author name exceeds maximum 256 characters")
	ErrEmbedTotalLengthTooLong = errors.New("embed total character count exceeds maximum 6000 characters")
)

type EmbedBuilder struct {
	embed types.Embed
	err   error
}

func NewEmbed() *EmbedBuilder {
	return &EmbedBuilder{
		embed: types.Embed{
			Type: "rich",
		},
	}
}

func (b *EmbedBuilder) SetTitle(title string) *EmbedBuilder {
	b.embed.Title = title
	return b
}

func (b *EmbedBuilder) SetDescription(desc string) *EmbedBuilder {
	b.embed.Description = desc
	return b
}

func (b *EmbedBuilder) SetURL(url string) *EmbedBuilder {
	b.embed.URL = url
	return b
}

func (b *EmbedBuilder) SetColor(color any) *EmbedBuilder {
	col, err := ResolveColor(color)
	if err != nil {
		b.err = err
		return b
	}
	b.embed.Color = col
	return b
}

func (b *EmbedBuilder) SetTimestamp(t any) *EmbedBuilder {
	if t == nil {
		b.embed.Timestamp = nil
		return b
	}
	switch val := t.(type) {
	case time.Time:
		utc := val.UTC()
		b.embed.Timestamp = &utc
	case *time.Time:
		if val != nil {
			utc := val.UTC()
			b.embed.Timestamp = &utc
		} else {
			b.embed.Timestamp = nil
		}
	default:
		item := util.New(val)
		if item.Error() != nil {
			b.err = item.Error()
			return b
		}
		tm := item.Time()
		b.embed.Timestamp = &tm
	}
	return b
}

func (b *EmbedBuilder) SetFooter(text string, iconURL ...string) *EmbedBuilder {
	var icon string
	if len(iconURL) > 0 {
		icon = iconURL[0]
	}
	b.embed.Footer = &types.EmbedFooter{
		Text:    text,
		IconURL: icon,
	}
	return b
}

func (b *EmbedBuilder) SetAuthor(name string, urls ...string) *EmbedBuilder {
	var url, icon string
	if len(urls) > 0 {
		url = urls[0]
	}
	if len(urls) > 1 {
		icon = urls[1]
	}
	b.embed.Author = &types.EmbedAuthor{
		Name:    name,
		URL:     url,
		IconURL: icon,
	}
	return b
}

func (b *EmbedBuilder) SetImage(url string) *EmbedBuilder {
	b.embed.Image = &types.EmbedMedia{
		URL: url,
	}
	return b
}

func (b *EmbedBuilder) SetThumbnail(url string) *EmbedBuilder {
	b.embed.Thumbnail = &types.EmbedMedia{
		URL: url,
	}
	return b
}

func (b *EmbedBuilder) AddField(name, value string, inline ...bool) *EmbedBuilder {
	isInline := false
	if len(inline) > 0 {
		isInline = inline[0]
	}
	b.embed.Fields = append(b.embed.Fields, types.EmbedField{
		Name:   name,
		Value:  value,
		Inline: isInline,
	})
	return b
}

func (b *EmbedBuilder) AddFields(fields ...types.EmbedField) *EmbedBuilder {
	b.embed.Fields = append(b.embed.Fields, fields...)
	return b
}

func (b *EmbedBuilder) ClearFields() *EmbedBuilder {
	b.embed.Fields = nil
	return b
}

func (b *EmbedBuilder) ToEmbed() types.Embed {
	return b.embed
}

func (b *EmbedBuilder) Build() (*types.Embed, error) {
	if b.err != nil {
		return nil, b.err
	}

	totalLen := 0

	titleLen := utf8.RuneCountInString(b.embed.Title)
	if titleLen > 256 {
		return nil, ErrEmbedTitleTooLong
	}
	totalLen += titleLen

	descLen := utf8.RuneCountInString(b.embed.Description)
	if descLen > 4096 {
		return nil, ErrEmbedDescriptionTooLong
	}
	totalLen += descLen

	if len(b.embed.Fields) > 25 {
		return nil, ErrEmbedTooManyFields
	}

	for _, f := range b.embed.Fields {
		if strings.TrimSpace(f.Name) == "" {
			return nil, ErrEmbedFieldMissingName
		}
		nameLen := utf8.RuneCountInString(f.Name)
		if nameLen > 256 {
			return nil, ErrEmbedFieldNameTooLong
		}
		if strings.TrimSpace(f.Value) == "" {
			return nil, ErrEmbedFieldMissingValue
		}
		valLen := utf8.RuneCountInString(f.Value)
		if valLen > 1024 {
			return nil, ErrEmbedFieldValueTooLong
		}
		totalLen += nameLen + valLen
	}

	if b.embed.Footer != nil {
		footerLen := utf8.RuneCountInString(b.embed.Footer.Text)
		if footerLen > 2048 {
			return nil, ErrEmbedFooterTooLong
		}
		totalLen += footerLen
	}

	if b.embed.Author != nil {
		authorLen := utf8.RuneCountInString(b.embed.Author.Name)
		if authorLen > 256 {
			return nil, ErrEmbedAuthorTooLong
		}
		totalLen += authorLen
	}

	if totalLen > 6000 {
		return nil, ErrEmbedTotalLengthTooLong
	}

	res := b.embed
	return &res, nil
}
