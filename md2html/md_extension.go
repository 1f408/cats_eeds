package md2html

import (
	cmhtml "github.com/alecthomas/chroma/v2/formatters/html"
	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	emoji_ast "github.com/yuin/goldmark-emoji/ast"
	emoji_def "github.com/yuin/goldmark-emoji/definition"
	highlight "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/util"

	alerts "github.com/1f408/cats_eeds/md2html/alerts"
	md_embed "github.com/1f408/cats_eeds/md2html/embed"
	tasklist "github.com/1f408/cats_eeds/md2html/tasklist"
)

func NewParserExts(mc *MdConfig) []goldmark.Extender {
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
			emoji.New(
				emoji.WithEmojis(em_list),
				emoji.WithRenderingMethod(emoji.Func),
				emoji.WithRendererFunc(func(w util.BufWriter, _ []byte, n *emoji_ast.Emoji, _ *emoji.RendererConfig) {
					w.WriteString(string(n.Value.Unicode))
				}),
			))
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

	if mc.Extension.Alerts {
		parser_exts = append(parser_exts, alerts.NewAlertBlock(
			alerts.WithTitleHtmlMaping(alerts.TitleHtmlMapping(*mc.Alerts.TitleMapping.Value)),
		))
	}

	return parser_exts
}
