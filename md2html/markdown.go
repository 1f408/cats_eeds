package md2html

import (
	_ "embed"
	"io/fs"
	"time"

	"github.com/l4go/recode"
	"github.com/naoina/toml"

	"github.com/1f408/cats_eeds/upath"
)

//go:embed markdown.conf
var defaultMdConfig []byte

type ExtFlags struct {
	Table          bool `toml:",omitempty"`
	Strikethrough  bool `toml:",omitempty"`
	TaskList       bool `toml:",omitempty"`
	DefinitionList bool `toml:",omitempty"`
	Footnote       bool `toml:",omitempty"`
	Autolinks      bool `toml:",omitempty"`
	Cjk            bool `toml:",omitempty"`
	Emoji          bool `toml:",omitempty"`
	Highlight      bool `toml:",omitempty"`
	Math           bool `toml:",omitempty"`
	Mermaid        bool `toml:",omitempty"`
	GeoMap         bool `toml:",omitempty"`
	Embed          bool `toml:",omitempty"`
	Alerts         bool `toml:",omitempty"`
	MsInclude      bool `toml:",omitempty"`
	DataTable      bool `toml:",omitempty"`
}

type AutoIdsOptions struct {
	Type string `toml:",omitempty"`
}

type FootnoteOptions struct {
	BacklinkHTML string `toml:",omitempty"`
}

type EmojiOptions struct {
	Mapping upath.Import[*EmojiMapping] `toml:",omitempty"`
}

type EmbedOptions struct {
	Rules upath.Import[*EmbedRules] `toml:",omitempty"`
}

type AlertsOptions struct {
	TitleMapping upath.Import[*AlertTitleMapping] `toml:",omitempty"`
}

type MdConfig struct {
	Extension ExtFlags
	AutoIds   AutoIdsOptions  `toml:",omitempty"`
	Footnote  FootnoteOptions `toml:",omitempty"`
	Emoji     EmojiOptions    `toml:",omitempty"`
	Embed     EmbedOptions    `toml:",omitempty"`
	Alerts    AlertsOptions   `toml:",omitempty"`

	ModTime time.Time `toml:"-"`
	init    bool      `toml:"-"`
}

func NewMdConfigDefault() *MdConfig {
	type rawMdConfig MdConfig

	raw := rawMdConfig{}
	if err := toml.Unmarshal(defaultMdConfig, &raw); err != nil {
		panic("bad default MdConfig: " + err.Error())
	}

	mc := MdConfig(raw)
	return &mc
}

func (mc *MdConfig) Initialize() {
	type Raw MdConfig
	err := toml.Unmarshal(defaultMdConfig, (*Raw)(mc))
	if err != nil {
		panic("bad default markdown config file: " + err.Error())
	}

	mc.init = true
}

func (_ *MdConfig) MakeNew() *MdConfig {
	mc := &MdConfig{}
	mc.Initialize()

	return mc
}

func (mc *MdConfig) UnmarshalTOML(decode func(interface{}) error) error {
	if !mc.init {
		mc.Initialize()
	}

	type rawMdConfig MdConfig
	if err := decode((*rawMdConfig)(mc)); err != nil {
		return err
	}

	return nil
}

func NewMdConfig(fsys fs.FS, name string) (*MdConfig, error) {
	var md_cfg upath.Import[*MdConfig]

	p, err := upath.New(name)
	if err != nil {
		return nil, err
	}
	md_cfg.UPath = p
	if rerr := recode.RecursiveRebuild(&md_cfg, fsys); rerr != nil {
		return nil, rerr
	}

	return md_cfg.Value, nil
}
