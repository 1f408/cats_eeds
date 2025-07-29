package mdview

import (
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/l4go/rpath"

	"github.com/1f408/cats_eeds/md2html"
	"github.com/1f408/cats_eeds/upath"
	"github.com/1f408/cats_eeds/view/internal/dirview"
	"github.com/1f408/cats_eeds/view/internal/htpath"
	"github.com/1f408/cats_eeds/view/internal/mtable"
	"github.com/1f408/cats_eeds/view/internal/tmplext"
)

func is_abs_dir_path(p string) bool {
	return rpath.Clean(p) == p && rpath.IsDir(p) && len(p) > 0 && p[0] == '/'
}

func new_err(format string, v ...interface{}) error {
	return errors.New(fmt.Sprintf(format, v...))
}

type MdView struct {
	SystemFS fs.FS

	SocketType string
	SocketPath string

	CacheControl         string
	UrlTopPath           string
	UrlLibPath           string
	DirectoryRedirection bool

	DocumentRoot upath.UPath

	IndexName    string
	OriginTmpl   *template.Template
	HtmlTmpl     *template.Template
	MainTmplName string
	SvgIconPath  upath.UPath
	Md2Html      *md2html.Md2Html

	MimeExtTable      *mtable.MimeExtTable
	MarkdownExt       []string
	MarkdownConfig    *md2html.MdConfig
	CustomPageConfig  *md2html.CustomPageConfig
	PrintPaperMapping *md2html.PrintPaperMapping

	ThemeStyle   string
	PageStyle    string
	PrintSizeCss string
	PrintZoom    float32

	LocationNavi string
	TocNavi      string

	DirectoryViewMode       string
	DirectoryViewRoots      []upath.UPath
	DirectoryViewHidden     []*regexp.Regexp
	DirectoryViewPathHidden []*regexp.Regexp
	TimeStampFormat         string
	DirViewStamp            *dirview.DirViewStamp

	TextViewMode string

	ConfigModTime time.Time
	TemplateTag   []byte
	SystemHtmlIds []string
}

func newMdViewDefault() *MdView {
	mdv := &MdView{}

	mdv.UrlTopPath = "/"
	mdv.UrlLibPath = "/"
	mdv.IndexName = "README.md"
	mdv.MainTmplName = "mdview.tmpl"
	mdv.MarkdownExt = []string{"md", "markdown"}
	mdv.ThemeStyle = "radio"
	mdv.PageStyle = ""
	mdv.PrintSizeCss = ""

	mdv.LocationNavi = "dirs"
	mdv.TocNavi = "details"
	mdv.DirectoryViewMode = "autoindex"

	mdv.TimeStampFormat = "%F %T"
	mdv.TextViewMode = "html"

	return mdv
}

func NewMdView(cfg *MdViewConfig) (*MdView, error) {
	mdv := newMdViewDefault()

	mdv.SystemFS = cfg.SystemFS
	mdv.ConfigModTime = cfg.ModTime

	mdv.SocketType = cfg.SocketType
	mdv.SocketPath = cfg.SocketPath
	if mdv.SocketType != "tcp" && mdv.SocketType != "unix" {
		return nil, new_err("Bad socket type: %s", mdv.SocketType)
	}

	mdv.DirectoryRedirection = cfg.DirectoryRedirection
	mdv.CacheControl = cfg.CacheControl

	if cfg.UrlTopPath != "" {
		mdv.UrlTopPath = cfg.UrlTopPath
	}
	if !is_abs_dir_path(mdv.UrlTopPath) {
		return nil, new_err("Bad url_top_path directory: %s", mdv.UrlTopPath)
	}
	if cfg.UrlLibPath != "" {
		mdv.UrlLibPath = cfg.UrlLibPath
	}
	if !is_abs_dir_path(mdv.UrlLibPath) {
		return nil, new_err("Bad url_lib_path directory: %s", mdv.UrlLibPath)
	}

	if !cfg.DocumentRoot.IsZero() {
		mdv.DocumentRoot = cfg.DocumentRoot
	}
	if mdv.DocumentRoot.IsZero() {
		return nil, new_err("Must document root")
	}
	if fi, err := mdv.DocumentRoot.Stat(mdv.SystemFS); err != nil || !fi.IsDir() {
		return nil, new_err("Not found document root dirctory: %s", mdv.DocumentRoot.String())
	}

	if cfg.IndexName != "" {
		mdv.IndexName = cfg.IndexName
	}
	if strings.IndexRune(mdv.IndexName, '/') >= 0 {
		return nil, new_err("Bad index name")
	}

	if !cfg.IconPath.IsZero() {
		mdv.SvgIconPath = cfg.IconPath
	}
	if !mdv.SvgIconPath.IsZero() {
		if fi, err := mdv.SvgIconPath.Stat(mdv.SystemFS); err != nil || !fi.IsDir() {
			return nil, new_err("Not found ICON SVG dirctory: %s", mdv.SvgIconPath.String())
		}
	}

	mdv.DirectoryViewRoots = []upath.UPath{mdv.DocumentRoot}
	if cfg.DirectoryViewRoots != nil {
		mdv.DirectoryViewRoots = cfg.DirectoryViewRoots
	}

	if cfg.DirectoryViewHidden != nil {
		mdv.DirectoryViewHidden = make(
			[]*regexp.Regexp, len(cfg.DirectoryViewHidden))

		for i, ign_str := range cfg.DirectoryViewHidden {
			re, err := regexp.Compile(ign_str)
			if err != nil {
				return nil, new_err("Bad hidden pattern: %s", ign_str)
			}
			mdv.DirectoryViewHidden[i] = re
		}
	}
	if cfg.DirectoryViewPathHidden != nil {
		mdv.DirectoryViewPathHidden = make(
			[]*regexp.Regexp, len(cfg.DirectoryViewPathHidden))

		for i, ign_str := range cfg.DirectoryViewPathHidden {
			re, err := regexp.Compile(ign_str)
			if err != nil {
				return nil, new_err("Bad path hidden pattern: %s", ign_str)
			}
			mdv.DirectoryViewPathHidden[i] = re
		}
	}

	if cfg.DirectoryViewMode != "" {
		mdv.DirectoryViewMode = cfg.DirectoryViewMode
	}
	switch mdv.DirectoryViewMode {
	case "none":
	case "autoindex":
	case "close":
	case "auto":
	case "open":
	default:
		return nil, new_err("Bad directory view mode: %s", mdv.DirectoryViewMode)
	}

	if cfg.TimeStampFormat != "" {
		mdv.TimeStampFormat = cfg.TimeStampFormat
	}
	var err error
	mdv.DirViewStamp, err = dirview.NewDirViewStamp(
		mdv.SystemFS,
		mdv.DirectoryViewRoots, mdv.TimeStampFormat,
		mdv.DirectoryViewHidden, mdv.DirectoryViewPathHidden)
	if err != nil {
		return nil, new_err("Bad timestamp format: %s", mdv.TimeStampFormat)
	}

	if cfg.TextViewMode != "" {
		mdv.TextViewMode = cfg.TextViewMode
	}
	switch mdv.TextViewMode {
	case "raw":
	case "html":
	default:
		return nil, new_err("Bad text view mode: %s", mdv.TextViewMode)
	}

	mdv.OriginTmpl = template.New("")
	tmpl_funcs := template.FuncMap{}
	tmplext.AddDefaultFunc(tmpl_funcs, mdv.SystemFS, mdv.SvgIconPath)

	mdv.OriginTmpl = mdv.OriginTmpl.Funcs(tmpl_funcs)
	mdv.OriginTmpl, err = mdv.OriginTmpl.ParseFS(mdv.SystemFS, upath.FSPaths(cfg.TmplPaths)...)

	if err != nil {
		return nil, new_err("Template parse error: %s", err)
	}

	if cfg.MainTmpl != "" {
		mdv.MainTmplName = cfg.MainTmpl
	}

	if len(*cfg.MimeExtTable.Value) != 0 {
		mdv.MimeExtTable = cfg.MimeExtTable.Value
	}
	for ext, mtype := range *mdv.MimeExtTable {
		if e := htpath.SetExtMimeType(ext, mtype); e != nil {
			return nil, new_err("Bad mime extension: %s", ext)
		}
	}

	if cfg.MarkdownExt != nil {
		mdv.MarkdownExt = cfg.MarkdownExt
	}
	for _, ext := range mdv.MarkdownExt {
		if e := htpath.SetMarkdownExt(ext); e != nil {
			return nil, new_err("Bad file extension: %s", ext)
		}
	}

	mdv.MarkdownConfig = cfg.MarkdownConfig.Value

	mdv.CustomPageConfig = cfg.CustomPageConfig.Value
	mdv.PrintPaperMapping = mdv.CustomPageConfig.PrintPaper.Mapping.Value

	if cfg.ThemeStyle != "" {
		mdv.ThemeStyle = cfg.ThemeStyle
	}
	switch mdv.ThemeStyle {
	case "radio":
	case "load":
	case "os":
	default:
		return nil, new_err("Bad ThemeStyle: %s", mdv.ThemeStyle)
	}

	if cfg.PageStyle != "" {
		mdv.PageStyle = cfg.PageStyle
	}
	if strings.ContainsRune(mdv.PageStyle, '/') {
		return nil, new_err("Bad page style: %s", mdv.PageStyle)
	}

	print_paper := mdv.CustomPageConfig.PrintPaper.Default.PaperType
	if print_paper != "" {
		css, ok := mdv.PrintPaperMapping.GetCss(print_paper)
		if !ok {
			return nil, new_err("Bad page size: %s", print_paper)
		}
		mdv.PrintSizeCss = css
	}

	print_zoom := mdv.CustomPageConfig.PrintPaper.Default.PrintZoom
	if print_zoom != 0.0 {
		mdv.PrintZoom = print_zoom
	}
	if mdv.PrintZoom < 0 {
		return nil, new_err("Bad print zoom: %f", mdv.PrintZoom)
	}

	if cfg.LocationNavi != "" {
		mdv.LocationNavi = cfg.LocationNavi
	}
	switch mdv.LocationNavi {
	case "none":
	case "dirs":
	default:
		return nil, new_err("Bad location navi type: %s", mdv.LocationNavi)
	}

	if cfg.TocNavi != "" {
		mdv.TocNavi = cfg.TocNavi
	}
	switch mdv.TocNavi {
	case "none":
	case "static":
	case "details":
	default:
		return nil, new_err("Bad toc navi type: %s", mdv.TocNavi)
	}

	sum, err := mdv.SumTemplate()
	if err != nil {
		return nil, new_err("Template execute error: %s", err)
	}

	mdv.TemplateTag = sum.Sha256
	mdv.SystemHtmlIds = sum.SystemIds

	mdv.Md2Html = md2html.NewMd2Html(&md2html.Md2HtmlConfig{
		MdConfig:  mdv.MarkdownConfig,
		SystemIds: mdv.SystemHtmlIds,
	})

	return mdv, nil
}
