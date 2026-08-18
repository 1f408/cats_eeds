package ms_include

import (
	"io"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type ConvertHtmlFunc func(abs_file string, w io.Writer) error

type msIncludeExtension struct{
	convertHtml ConvertHtmlFunc
	pathStack PathStack
}

func NewMsInclude(conv ConvertHtmlFunc, pstack PathStack) goldmark.Extender {
	return &msIncludeExtension{
		convertHtml: conv,
		pathStack: pstack,
	}
}

func (e *msIncludeExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(NewMsIncludeParser(), 150),
		),
	)

	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(NewIncludeHTMLRenderer(e.convertHtml, e.pathStack), 500),
		),
	)
}
