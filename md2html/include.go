package md2html

import (
	"errors"
	"io"
	"io/fs"

	"github.com/l4go/rpath"
	"github.com/l4go/unifs"

	"github.com/1f408/cats_eeds/frontmatter"
	"github.com/1f408/cats_eeds/internal/ftype"
)

var ErrNoFMParam = errors.New("not found")
var ErrNotMarkdown = errors.New("not markdown")

type ConvertFuncParam struct {
	Md2Html     *Md2Html
	SystemFS    fs.FS
	FrontMatter FrontMatterConfig
}

func (cf *ConvertFuncParam) ConvertHtml(fs_file string, w io.Writer) error {
	var err error
	fs_file, err = unifs.Clean(fs_file)
	if err != nil {
		return err
	}
	fs_dir := rpath.Dir(fs_file)

	if kind, _ := ftype.GetFileKindByExt(rpath.Ext(fs_file)); kind != "text/markdown" {
		return ErrNotMarkdown
	}

	sysfs := cf.SystemFS
	fm := cf.FrontMatter

	m2h := cf.Md2Html

	raw_bin, rd_err := unifs.ReadFile(sysfs, fs_file)
	if rd_err != nil {
		return rd_err
	}

	md_doc := raw_bin
	if fm.IsEnabled() {
		body, fm_param, fm_err := fm.TrimAndParse(raw_bin)
		switch fm_err {
		case nil:
			if fm_param == nil {
				panic("nil FrontMatterParam")
			}
			md_doc = body
		case frontmatter.ErrNotFound:
		default:
			return fm_err
		}

		if fm_param != nil && fm_param.MarkdownConfig != "" {
			name := fm_param.MarkdownConfig
			if name[0] != '/' {
				name = rpath.Join(fs_dir, name)
			}
			if md_cfg, err := NewMdConfig(sysfs, name); err == nil {
				m2h = m2h.NewLocalSpec(md_cfg)
			}
		}
	}

	doc_bin := m2h.md2html(md_doc)
	_, werr := w.Write(doc_bin)
	return werr
}
