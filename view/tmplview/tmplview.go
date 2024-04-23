package tmplview

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/l4go/rpath"

	"github.com/1f408/cats_eeds/authz"
	"github.com/1f408/cats_eeds/md2html"
	"github.com/1f408/cats_eeds/upath"

	"github.com/1f408/cats_eeds/view/internal/dirview"
	"github.com/1f408/cats_eeds/view/internal/htpath"
	"github.com/1f408/cats_eeds/view/internal/mtable"
)

func new_err(format string, v ...interface{}) error {
	return errors.New(fmt.Sprintf(format, v...))
}

type TmplView struct {
	SocketType string
	SocketPath string

	SystemFS fs.FS

	CacheControl string
	UrlTopPath   string
	UrlLibPath   string

	UserMap         *authz.UserMap
	AuthnUserHeader string

	OriginTmpl *template.Template

	DocumentRoot upath.UPath
	SvgIconPath  upath.UPath
	IndexName    string

	MdTmplName string
	Md2Html    *md2html.Md2Html

	MimeExtTable   *mtable.MimeExtTable
	MarkdownExt    []string
	MarkdownConfig *md2html.MdConfig

	ThemeStyle   string
	LocationNavi string

	DirectoryViewMode       string
	DirectoryViewRoots      []upath.UPath
	DirectoryViewHidden     []*regexp.Regexp
	DirectoryViewPathHidden []*regexp.Regexp
	TimeStampFormat         string
	DirViewStamp            *dirview.DirViewStamp

	TextViewMode string

	CatUiConfigPath upath.UPath
	CatUiConfigExt  string
	CatUiTmplName   string

	ConfigModTime time.Time
	TemplateTag   []byte
}

func newTmplViewDefault() *TmplView {
	tmpv := &TmplView{}

	tmpv.UrlTopPath = "/"
	tmpv.UrlLibPath = "/"

	tmpv.AuthnUserHeader = "X-Forwarded-User"
	tmpv.IndexName = "README.md"
	tmpv.MdTmplName = "mdview.tmpl"
	tmpv.MarkdownExt = []string{"md", "markdown"}

	tmpv.ThemeStyle = "radio"
	tmpv.LocationNavi = "dirs"

	tmpv.DirectoryViewMode = "autoindex"
	tmpv.TimeStampFormat = "%F %T"

	tmpv.TextViewMode = "html"

	tmpv.CatUiConfigExt = "ui"

	return tmpv
}

func (tmpv *TmplView) Warn(format string, v ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", v...)
}

var DumpPath string = ""

func NewTmplView(cfg *TmplViewConfig) (*TmplView, error) {
	tmpv := newTmplViewDefault()

	tmpv.SystemFS = cfg.SystemFS
	tmpv.ConfigModTime = cfg.ModTime

	tmpv.SocketType = cfg.SocketType
	tmpv.SocketPath = cfg.SocketPath
	if tmpv.SocketType != "tcp" && tmpv.SocketType != "unix" {
		return nil, new_err("Bad socket type: %s", tmpv.SocketType)
	}

	tmpv.CacheControl = cfg.CacheControl

	if cfg.UrlTopPath != "" {
		tmpv.UrlTopPath = rpath.SetDir("/" + cfg.UrlTopPath)
	}
	if cfg.UrlLibPath != "" {
		tmpv.UrlLibPath = rpath.SetDir("/" + cfg.UrlLibPath)
	}

	if cfg.Authz.AuthnUserHeader != "" {
		tmpv.AuthnUserHeader = cfg.Authz.AuthnUserHeader
	}

	var err error
	var user_map_cfg *authz.UserMapConfig
	user_map_cfg, err = authz.NewUserMapConfigFS(tmpv.SystemFS, cfg.Authz.UserMapConfig)
	if err != nil {
		return nil, new_err("user map config parse error: %s: %s",
			cfg.Authz.UserMapConfig, err)
	}

	tmpv.UserMap, err = authz.NewUserMapFS(tmpv.SystemFS, cfg.Authz.UserMap, user_map_cfg)
	if err != nil {
		return nil, new_err("user map parse error: %s: %s", cfg.Authz.UserMap, err)
	}

	if !cfg.Tmpl.IconPath.IsZero() {
		tmpv.SvgIconPath = cfg.Tmpl.IconPath
	}
	if !tmpv.SvgIconPath.IsZero() {
		if fi, err := tmpv.SvgIconPath.Stat(tmpv.SystemFS); err != nil || !fi.IsDir() {
			return nil, new_err("Not found ICON SVG dirctory: %s", tmpv.SvgIconPath.String())
		}
	}

	if !cfg.Tmpl.DocumentRoot.IsZero() {
		tmpv.DocumentRoot = cfg.Tmpl.DocumentRoot
	}
	if tmpv.DocumentRoot.IsZero() {
		return nil, new_err("Must document root")
	}
	if fi, err := tmpv.DocumentRoot.Stat(tmpv.SystemFS); err != nil || !fi.IsDir() {
		return nil, new_err("Not found root dirctory: %s", tmpv.DocumentRoot.String())
	}

	if cfg.Tmpl.IndexName != "" {
		tmpv.IndexName = cfg.Tmpl.IndexName
	}
	if strings.IndexRune(tmpv.IndexName, '/') >= 0 {
		return nil, new_err("Bad index template name")
	}
	if cfg.Tmpl.MdTmplName != "" {
		tmpv.MdTmplName = cfg.Tmpl.MdTmplName
	}
	if strings.IndexRune(tmpv.MdTmplName, '/') >= 0 {
		return nil, new_err("Bad markdown html template name")
	}

	tmpv.DirectoryViewRoots = []upath.UPath{tmpv.DocumentRoot}
	if cfg.Tmpl.DirectoryViewRoots != nil {
		tmpv.DirectoryViewRoots = cfg.Tmpl.DirectoryViewRoots
	}

	if cfg.Tmpl.DirectoryViewHidden != nil {
		tmpv.DirectoryViewHidden = make(
			[]*regexp.Regexp, len(cfg.Tmpl.DirectoryViewHidden))

		for i, ign_str := range cfg.Tmpl.DirectoryViewHidden {
			re, err := regexp.Compile(ign_str)
			if err != nil {
				return nil, new_err("Bad direcotor ignore pattern: %s", ign_str)
			}
			tmpv.DirectoryViewHidden[i] = re
		}
	}
	if cfg.Tmpl.DirectoryViewPathHidden != nil {
		tmpv.DirectoryViewPathHidden = make(
			[]*regexp.Regexp, len(cfg.Tmpl.DirectoryViewPathHidden))

		for i, ign_str := range cfg.Tmpl.DirectoryViewPathHidden {
			re, err := regexp.Compile(ign_str)
			if err != nil {
				return nil, new_err("Bad direcotor ignore pattern: %s", ign_str)
			}
			tmpv.DirectoryViewPathHidden[i] = re
		}
	}

	if cfg.Tmpl.DirectoryViewMode != "" {
		tmpv.DirectoryViewMode = cfg.Tmpl.DirectoryViewMode
	}
	switch tmpv.DirectoryViewMode {
	case "none":
	case "autoindex":
	case "close":
	case "auto":
	case "open":
	default:
		return nil, new_err("Bad directory view mode: %s", tmpv.DirectoryViewMode)
	}

	if cfg.Tmpl.TimeStampFormat != "" {
		tmpv.TimeStampFormat = cfg.Tmpl.TimeStampFormat
	}
	tmpv.DirViewStamp, err = dirview.NewDirViewStamp(
		tmpv.SystemFS,
		tmpv.DirectoryViewRoots, tmpv.TimeStampFormat,
		tmpv.DirectoryViewHidden, tmpv.DirectoryViewPathHidden)
	if err != nil {
		return nil, new_err("Bad timestamp format: %s", tmpv.TimeStampFormat)
	}

	if cfg.Tmpl.TextViewMode != "" {
		tmpv.TextViewMode = cfg.Tmpl.TextViewMode
	}
	switch tmpv.TextViewMode {
	case "raw":
	case "html":
	default:
		return nil, new_err("Bad text view mode: %s", tmpv.TextViewMode)
	}

	if !cfg.Tmpl.CatUiConfigPath.IsZero() {
		tmpv.CatUiConfigPath = cfg.Tmpl.CatUiConfigPath
	}
	if !tmpv.CatUiConfigPath.IsZero() {
		if fi, err := tmpv.CatUiConfigPath.Stat(tmpv.SystemFS); err != nil || !fi.IsDir() {
			return nil, new_err("Not found Cat UI config dirctory: %s",
				tmpv.CatUiConfigPath.String())
		}
	}

	if cfg.Tmpl.CatUiConfigExt != "" {
		tmpv.CatUiConfigExt = cfg.Tmpl.CatUiConfigExt
	}
	if cfg.Tmpl.CatUiTmplName != "" {
		tmpv.CatUiTmplName = cfg.Tmpl.CatUiTmplName
	}

	tmpv.OriginTmpl = template.New("")
	tmpl_funcs := template.FuncMap{
		"once":      DummyTmplOnce,
		"svg_icon":  tmpv.TmplSvgIcon,
		"file_type": func(s string) string { return "" },
		"in_group":  func(grp string) bool { return false },
		"in_user":   func() bool { return false },
		"cat_ui":    tmpv.CatUi,
	}
	tmpv.OriginTmpl = tmpv.OriginTmpl.Funcs(tmpl_funcs)

	tmpv.OriginTmpl, err = tmpv.OriginTmpl.ParseFS(
		tmpv.SystemFS, upath.FSPaths(cfg.Tmpl.TmplPaths)...)
	if err != nil {
		return nil, new_err("Template parse error: %s", err)
	}

	if len(*cfg.Tmpl.MimeExtTable.Value) != 0 {
		tmpv.MimeExtTable = cfg.Tmpl.MimeExtTable.Value
	}
	for ext, mtype := range *tmpv.MimeExtTable {
		if e := htpath.SetExtMimeType(ext, mtype); e != nil {
			return nil, new_err("Bad mime extension: %s", ext)
		}
	}
	if cfg.Tmpl.MarkdownExt != nil {
		tmpv.MarkdownExt = cfg.Tmpl.MarkdownExt
	}
	for _, ext := range tmpv.MarkdownExt {
		if e := htpath.SetMarkdownExt(ext); e != nil {
			return nil, new_err("Bad file extension: %s", ext)
		}
	}

	tmpv.MarkdownConfig = cfg.Tmpl.MarkdownConfig.Value
	tmpv.Md2Html = md2html.NewMd2Html(tmpv.MarkdownConfig)

	if cfg.Tmpl.ThemeStyle != "" {
		tmpv.ThemeStyle = cfg.Tmpl.ThemeStyle
	}
	switch tmpv.ThemeStyle {
	case "radio":
	case "os":
	default:
		return nil, new_err("Bad ThemeStyle: %s", tmpv.ThemeStyle)
	}

	if cfg.Tmpl.LocationNavi != "" {
		tmpv.LocationNavi = cfg.Tmpl.LocationNavi
	}
	switch tmpv.LocationNavi {
	case "none":
	case "dirs":
	default:
		return nil, new_err("Bad location navi type: %s", tmpv.LocationNavi)
	}

	sum, err := tmpv.SumTemplate()
	if err != nil {
		return nil, new_err("Template execute error: %s", err)
	}

	tmpv.TemplateTag = sum

	return tmpv, nil
}
