package alerts

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"

	"github.com/1f408/cats_eeds/md2html/htfix"
)

type TitleHtmlMapping map[string]string
type Config struct {
	TitleHtmlMapping TitleHtmlMapping
}

var DefaultConfig = Config{
	TitleHtmlMapping: TitleHtmlMapping{
		"NOTE":      htfix.FixHTMLString(`🔍 Note`),
		"TIP":       htfix.FixHTMLString(`💡 Tip`),
		"IMPORTANT": htfix.FixHTMLString(`🔔 Important`),
		"WARNING":   htfix.FixHTMLString(`📣 Warning`),
		"CAUTION":   htfix.FixHTMLString(`💣 Caution`),
	},
}

type Option interface {
	SetAlertBlockOption(*Config)
}

type withTitleMapping struct {
	value TitleHtmlMapping
}

func (o *withTitleMapping) SetAlertBlockOption(c *Config) {
	c.TitleHtmlMapping = o.value
}

func WithTitleHtmlMaping(v TitleHtmlMapping) Option {
	mp := make(TitleHtmlMapping, len(v))
	for k, v := range v {
		mp[k] = htfix.FixHTMLString(v)
	}
	return &withTitleMapping{value: mp}
}

func NewAlertBlock(opts ...Option) goldmark.Extender {
	return &alertBlock{
		options: opts,
	}
}

type alertBlock struct {
	options []Option
}

func (e *alertBlock) Extend(md goldmark.Markdown) {
	md.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(NewAlertBlockParser(e.options...), 750),
	))
	md.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewAlertBlockRenderer(e.options...), 500),
		util.Prioritized(NewAlertTitleRenderer(e.options...), 500),
	))
}
