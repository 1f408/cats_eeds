package mdview

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

type MdViewConfig struct {
	SocketType   string
	SocketPath   string
	CacheControl string `toml:",omitempty"`

	UrlTopPath string `toml:",omitempty"`
	UrlLibPath string `toml:",omitempty"`

	DocumentRoot upath.UPath
	IndexName    string `toml:",omitempty"`
	TmplPaths    []upath.UPath
	IconPath     upath.UPath `toml:",omitempty"`
	MainTmpl     string      `toml:",omitempty"`

	MimeExtTable   upath.Import[*mtable.MimeExtTable] `toml:",omitempty"`
	MarkdownExt    []string                           `toml:",omitempty"`
	MarkdownConfig upath.Import[*md2html.MdConfig]    `toml:",omitempty"`

	ThemeStyle   string `toml:",omitempty"`
	LocationNavi string `toml:",omitempty"`

	DirectoryViewMode       string
	DirectoryViewRoots      []upath.UPath `toml:",omitempty"`
	DirectoryViewHidden     []string      `toml:",omitempty"`
	DirectoryViewPathHidden []string      `toml:",omitempty"`
	TimeStampFormat         string        `toml:",omitempty"`

	TextViewMode string `toml:",omitempty"`

	SystemFS fs.FS     `toml:"-"`
	ModTime  time.Time `toml:"-"`
}

func NewMdViewConfig(file string) (*MdViewConfig, error) {
	return NewMdViewConfigFS(osfs.OsRootFS, file)
}

func NewMdViewConfigFS(fsys fs.FS, file string) (*MdViewConfig, error) {
	cfg_f, err := unifs.Open(fsys, file)
	if err != nil {
		return nil, new_err("Config file open error: %s: %s", file, err)
	}
	defer cfg_f.Close()
	text, err := io.ReadAll(cfg_f)
	if err != nil {
		return nil, new_err("Config read error: %s: %s", file, err)
	}

	cfg := &MdViewConfig{}
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
	if cfg.ModTime.Before(cfg.MarkdownConfig.Value.ModTime) {
		cfg.ModTime = cfg.MarkdownConfig.Value.ModTime
	}

	return cfg, nil
}
