package tmplview

import (
	"io"
	"io/fs"
	"time"

	"github.com/l4go/osfs"
	"github.com/l4go/recode"
	"github.com/l4go/unifs"
	"github.com/naoina/toml"

	"github.com/1f408/cats_eeds/md2html"
	"github.com/1f408/cats_eeds/upath"
	"github.com/1f408/cats_eeds/view/internal/mtable"
)

type TmplViewConfig struct {
	SocketType   string
	SocketPath   string
	CacheControl string `toml:",omitempty"`

	UrlTopPath string `toml:",omitempty"`
	UrlLibPath string `toml:",omitempty"`

	Authz authzConfig
	Tmpl  tmplConfig

	SystemFS fs.FS     `toml:"-"`
	ModTime  time.Time `toml:"-"`
}

type authzConfig struct {
	UserMapConfig   string `toml:",omitempty"`
	UserMap         string
	AuthnUserHeader string `toml:",omitempty"`
}
type tmplConfig struct {
	DocumentRoot upath.UPath
	IndexName    string `toml:",omitempty"`
	TmplPaths    []upath.UPath
	IconPath     upath.UPath `toml:",omitempty"`
	MdTmplName   string      `toml:",omitempty"`

	MimeExtTable   upath.Import[*mtable.MimeExtTable] `toml:",omitempty"`
	MarkdownExt    []string                           `toml:",omitempty"`
	MarkdownConfig upath.Import[*md2html.MdConfig]    `toml:",omitempty"`

	ThemeStyle   string `toml:",omitempty"`
	LocationNavi string `toml:",omitempty"`

	DirectoryViewMode       string        `toml:",omitempty"`
	DirectoryViewRoots      []upath.UPath `toml:",omitempty"`
	DirectoryViewHidden     []string      `toml:",omitempty"`
	DirectoryViewPathHidden []string      `toml:",omitempty"`
	TimeStampFormat         string        `toml:",omitempty"`

	TextViewMode string `toml:",omitempty"`

	CatUiConfigPath upath.UPath `toml:",omitempty"`
	CatUiConfigExt  string      `toml:",omitempty"`
	CatUiTmplName   string      `toml:",omitempty"`
}

func NewTmplViewConfig(file string) (*TmplViewConfig, error) {
	return NewTmplViewConfigFS(osfs.OsRootFS, file)
}

func NewTmplViewConfigFS(fsys fs.FS, file string) (*TmplViewConfig, error) {
	cfg_f, err := unifs.Open(fsys, file)
	if err != nil {
		return nil, new_err("Config file open error: %s: %s", file, err)
	}
	defer cfg_f.Close()

	text, err := io.ReadAll(cfg_f)
	if err != nil {
		return nil, new_err("Config read error: %s: %s", file, err)
	}

	cfg := &TmplViewConfig{}
	if err := toml.Unmarshal(text, cfg); err != nil {
		return nil, new_err("Config file parse error: %s: %s", file, err)
	}
	if err := recode.RecursiveRebuild(cfg, fsys); err != nil {
		return nil, new_err("Config file import error: %s", err)
	}

	cfg.SystemFS = fsys
	if fi, err := cfg_f.Stat(); err == nil {
		cfg.ModTime = fi.ModTime()
	}
	if cfg.ModTime.Before(cfg.Tmpl.MarkdownConfig.ModTime) {
		cfg.ModTime = cfg.Tmpl.MarkdownConfig.ModTime
	}

	return cfg, nil
}
