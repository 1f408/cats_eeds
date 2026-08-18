package md2html

import (
	"bytes"
	"fmt"
	"io/fs"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/1f408/cats_eeds/md2html/ms_include"
	"github.com/1f408/cats_eeds/md2html/uniqid"
)

type Md2Html struct {
	cfg       *MdConfig
	sani      *sanitaizer
	sys_ids   []string
	md_parser goldmark.Markdown
	id_tbl    uniqid.IdsTable
	inc_cfg   *IncludeConfig
}

type Md2HtmlConfig struct {
	MdConfig    *MdConfig
	SystemIds   []string
	SystemFS    fs.FS
	FrontMatter FrontMatterConfig
	StartMdFile string
}

type IncludeConvertHtml = ms_include.ConvertHtmlFunc
type IncludePathStack = ms_include.PathStack

type IncludeConfig struct {
	ConvertHtml IncludeConvertHtml
	PathStack   IncludePathStack
}

func NewMd2Html(cfg *Md2HtmlConfig) *Md2Html {
	md_cfg := cfg.MdConfig
	if md_cfg == nil {
		md_cfg = NewMdConfigDefault()
	}
	id_tbl := uniqid.NewMapIdsTable()
	for _, id := range cfg.SystemIds {
		id_tbl.Put([]byte(id))
	}

	if cfg.SystemFS == nil {
		panic("MUST SystemFS")
	}
	if len(cfg.StartMdFile) == 0 || cfg.StartMdFile[0] != '/' {
		panic("MUST StartMdFile")
	}

	m2h := &Md2Html{
		cfg:     md_cfg,
		sani:    newSanitizer(),
		sys_ids: cfg.SystemIds,
		id_tbl:  id_tbl,
	}

	cf_pm := &ConvertFuncParam{
		Md2Html:     m2h,
		SystemFS:    cfg.SystemFS,
		FrontMatter: cfg.FrontMatter,
	}

	inc_cfg := &IncludeConfig{
		ConvertHtml: cf_pm.ConvertHtml,
		PathStack:   ms_include.NewSlicePathStack(cfg.StartMdFile),
	}
	m2h.inc_cfg = inc_cfg

	parser_exts := NewParserExts(md_cfg, id_tbl, inc_cfg)
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
	m2h.md_parser = md_parser

	return m2h
}

func (m2h *Md2Html) ResetIds() {
	id_tbl := uniqid.NewMapIdsTable()
	for _, id := range m2h.sys_ids {
		id_tbl.Put([]byte(id))
	}
	m2h.id_tbl = id_tbl
}

func (m2h *Md2Html) NewLocalSpec(md_cfg *MdConfig) *Md2Html {
	parser_exts := NewParserExts(md_cfg, m2h.id_tbl, m2h.inc_cfg)

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
		cfg:       md_cfg,
		md_parser: md_parser,
		sani:      m2h.sani,
		sys_ids:   m2h.sys_ids,
		id_tbl:    m2h.id_tbl,
		inc_cfg:   m2h.inc_cfg,
	}
}

func (m2h *Md2Html) md2html(md []byte) []byte {
	var buf bytes.Buffer
	opts := []parser.ParseOption{}

	new_ids, err := NewAutoIds(m2h.md_parser, m2h.cfg.AutoIds.Type, m2h.id_tbl)
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
