package md2html

import (
	"bytes"
	"errors"
	"fmt"

	cmhtml "github.com/alecthomas/chroma/v2/formatters/html"
	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	emoji_def "github.com/yuin/goldmark-emoji/definition"
	highlight "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	md_embed "github.com/1f408/cats_eeds/md2html/embed"
	tasklist "github.com/1f408/cats_eeds/md2html/tasklist"
)

type Md2Html struct {
	cfg       *MdConfig
	md_parser goldmark.Markdown
	sani      *sanitaizer
}

var AutoIdsMap = map[string]func() parser.IDs{"": NewSafeIDs}

var ErrBadAutoIdsType = errors.New("bad auto IDs type")

func NewMd2Html(mc *MdConfig) *Md2Html {
	if mc == nil {
		*mc = *(NewMdConfigDefault())
	}

	parser_exts := []goldmark.Extender{}
	if mc.Extension.Table {
		parser_exts = append(parser_exts, extension.Table)
	}
	if mc.Extension.Strikethrough {
		parser_exts = append(parser_exts, extension.Strikethrough)
	}
	if mc.Extension.TaskList {
		parser_exts = append(parser_exts, tasklist.TaskList)
	}
	if mc.Extension.DefinitionList {
		parser_exts = append(parser_exts, extension.DefinitionList)
	}
	if mc.Extension.Footnote {
		fn_ext := []extension.FootnoteOption{}
		if mc.Footnote.BacklinkHTML != "" {
			fn_ext = append(fn_ext,
				extension.WithFootnoteBacklinkHTML(
					[]byte(mc.Footnote.BacklinkHTML)))
		}

		parser_exts = append(parser_exts, extension.NewFootnote(fn_ext...))
	}
	if mc.Extension.Emoji {
		em_list := emoji_def.NewEmojis()
		for k, v := range *mc.Emoji.Mapping.Value {
			em_list.Add(emoji_def.NewEmojis(emoji_def.NewEmoji(k, []rune(v.Emoji), v.Aliases...)))
		}

		parser_exts = append(parser_exts,
			emoji.New(emoji.WithEmojis(em_list)))
	}
	if mc.Extension.Cjk {
		parser_exts = append(parser_exts, extension.CJK)
	}
	if mc.Extension.Autolinks {
		parser_exts = append(parser_exts, extension.Linkify)
	}
	if mc.Extension.Highlight {
		parser_exts = append(parser_exts,
			highlight.NewHighlighting(
				highlight.WithFormatOptions(
					cmhtml.WithClasses(true),
					cmhtml.ClassPrefix("chrm-"),
				),
			))
	}
	if mc.Extension.Math {
		parser_exts = append(parser_exts, mathjax.NewMathJax(
			mathjax.WithInlineDelim("", ""),
			mathjax.WithBlockDelim("", "")))
	}
	if mc.Extension.Embed {
		vd_opts := []md_embed.VideoOptions{}
		for _, p := range mc.Embed.Rules.Value.Video {
			vd_opts = append(vd_opts,
				md_embed.VideoOptions{
					SiteId: p.SiteId,
					Host:   p.Host,
					Path:   p.Path,
					Regex:  p.Regex,
				})
		}
		ad_opts := []md_embed.AudioOptions{}
		for _, p := range mc.Embed.Rules.Value.Audio {
			ad_opts = append(ad_opts,
				md_embed.AudioOptions{
					SiteId: p.SiteId,
					Host:   p.Host,
					Path:   p.Path,
					Regex:  p.Regex,
				})
		}
		ifm_opts := []md_embed.IframeOptions{}
		for _, p := range mc.Embed.Rules.Value.Iframe {
			ifm_opts = append(ifm_opts,
				md_embed.IframeOptions{
					SiteId: p.SiteId,
					Host:   p.Host,
					Type:   p.Type,
					Path:   p.Path,
					Query:  p.Query,
					Regex:  p.Regex,
					Player: p.Player,
				})
		}

		parser_exts = append(parser_exts, md_embed.NewEmbed(
			md_embed.WithEmbedVideoExt(mc.Embed.Rules.Value.VideoExt),
			md_embed.WithEmbedAudioExt(mc.Embed.Rules.Value.AudioExt),
			md_embed.WithEmbedVideoUrl(vd_opts),
			md_embed.WithEmbedAudioUrl(ad_opts),
			md_embed.WithEmbedIframeUrl(ifm_opts),
		))
	}

	md_parser := goldmark.New(
		goldmark.WithExtensions(parser_exts...),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)

	return &Md2Html{
		cfg:       mc,
		md_parser: md_parser,
		sani:      newSanitizer(),
	}
}

func (m2h *Md2Html) md2html(md []byte) []byte {
	var buf bytes.Buffer
	opts := []parser.ParseOption{}

	new_func, ok := AutoIdsMap[m2h.cfg.AutoIds.Type]
	if ok {
		ctx := parser.NewContext(parser.WithIDs(new_func()))
		opts = append(opts, parser.WithContext(ctx))
	} else {
		panic(fmt.Errorf("Md2Html config error: %s", ErrBadAutoIdsType))
	}

	m2h.md_parser.Convert(md, &buf, opts...)

	return buf.Bytes()
}

func (m2h *Md2Html) sanitize(html []byte) ([]byte, error) {
	if m2h.sani == nil {
		return html, nil
	}
	return m2h.sani.Sanitize(html)
}

func (m2h *Md2Html) Convert(md []byte) ([]byte, []byte, []byte, error) {
	html_bin, err := m2h.sanitize(m2h.md2html(md))
	if err != nil {
		return nil, nil, nil, err
	}
	toc, _ := NewToc(html_bin)
	return html_bin, toc.ConvertHtml(), []byte(toc.Title), nil
}
