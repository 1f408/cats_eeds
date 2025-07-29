package md2html

import (
	"bytes"
	"fmt"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type Md2Html struct {
	cfg       *MdConfig
	md_parser goldmark.Markdown
	sani      *sanitaizer
	sys_ids   []string
}

type Md2HtmlConfig struct {
	MdConfig *MdConfig
	SystemIds []string
}

func NewMd2Html(cfg *Md2HtmlConfig) *Md2Html {
	mc := cfg.MdConfig
	if mc == nil {
		mc = NewMdConfigDefault()
	}
	sys_ids := cfg.SystemIds

	parser_exts := NewParserExts(mc)

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
		sys_ids: sys_ids,
	}
}

func (m2h *Md2Html) md2html(md []byte) []byte {
	var buf bytes.Buffer
	opts := []parser.ParseOption{}

	new_ids, err := NewAutoIds(m2h.md_parser, m2h.cfg.AutoIds.Type, m2h.sys_ids)
	if err != nil {
		panic(fmt.Errorf("Md2Html config error: %s", err))
	}

	ctx := parser.NewContext(parser.WithIDs(new_ids))
	opts = append(opts, parser.WithContext(ctx))

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

	var toc_bin []byte = nil
	var title_bin []byte = nil
	if toc, terr := NewToc(html_bin); terr == nil {
		title_bin = []byte(toc.Title)
		toc_bin, err = m2h.sanitize(toc.ConvertHtml())
		if err != nil {
			return nil, nil, nil, err
		}
	}
	return html_bin, toc_bin, title_bin, nil
}
