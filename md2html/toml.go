package md2html

import (
	_ "embed"
	"fmt"
	"io"
	"io/fs"
	"regexp"
	"time"

	"github.com/l4go/unifs"
	"github.com/naoina/toml"

	"github.com/1f408/cats_eeds/upath"
)

//go:embed "emoji_mapping.conf"
var defaultEmojiMapping []byte

//go:embed markdown.conf
var defaultMdConfig []byte

//go:embed embed_rules.conf
var defaultEmbedRules []byte

func loadTOML(fsys fs.FS, name string) ([]byte, error) {
	f, err := unifs.Open(fsys, name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return io.ReadAll(f)
}

type EmojiConfig struct {
	Emoji   string
	Aliases []string
	init    bool `toml:"-"`
}

type EmojiMapping map[string]*EmojiConfig

func (em *EmojiMapping) Initialize() {
	type Raw EmojiMapping
	if err := toml.Unmarshal(defaultEmojiMapping, (*Raw)(em)); err != nil {
		panic("bad default config file: " + err.Error())
	}
}

func (_ *EmojiMapping) MakeNew() *EmojiMapping {
	em := &EmojiMapping{}
	em.Initialize()

	return em
}

func (em *EmojiMapping) UnmarshalTOML(decode func(interface{}) error) error {
	if len(*em) == 0 {
		em.Initialize()
	}

	type Raw EmojiMapping
	return decode((*Raw)(em))
}

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

type EmbedRules struct {
	AudioExt []string    `toml:",omitempty"`
	VideoExt []string    `toml:",omitempty"`
	Video    []VideoOpt  `toml:",omitempty"`
	Audio    []AudioOpt  `toml:",omitempty"`
	Iframe   []IframeOpt `toml:",omitempty"`
	init     bool        `toml:"-"`
}

func (er *EmbedRules) IsZero() bool {
	return !er.init
}

func (er *EmbedRules) Initialize() {
	type Raw EmbedRules
	err := toml.Unmarshal(defaultEmbedRules, (*Raw)(er))
	if err != nil {
		panic("bad default config file: " + err.Error())
	}

	er.init = true
}

func (_ *EmbedRules) MakeNew() *EmbedRules {
	type Raw EmbedRules
	er := &EmbedRules{}
	er.Initialize()

	return er
}

func (er *EmbedRules) UnmarshalTOML(decode func(interface{}) error) error {
	if er.IsZero() {
		er.Initialize()
	}

	type Raw EmbedRules
	return decode((*Raw)(er))
}

type AudioOpt struct {
	SiteId string
	Host   string
	Path   string         `toml:",omitempty"`
	Regex  *regexp.Regexp `toml:",omitempty"`
}

func (ao *AudioOpt) UnmarshalTOML(decode func(interface{}) error) error {
	type rawAudioOpt struct {
		SiteId string
		Host   string
		Path   string `toml:",omitempty"`
		Regex  string `toml:",omitempty"`
	}

	var err error
	rao := rawAudioOpt{}
	if err = decode(&rao); err != nil {
		return err
	}

	var re *regexp.Regexp = nil
	if rao.Regex != "" {
		re, err = regexp.Compile(rao.Regex)
		if err != nil {
			return err
		}
	}

	ao.SiteId = rao.SiteId
	ao.Host = rao.Host
	ao.Path = rao.Path
	ao.Regex = re

	if ao.Host == "" {
		return fmt.Errorf("Missing 'host' parameter: %s", ao.SiteId)
	}
	if ao.Path == "" && ao.Regex == nil {
		return fmt.Errorf("Missing 'path' or 'regex' parameter: %s", ao.SiteId)
	}
	return nil
}

type VideoOpt struct {
	SiteId string
	Host   string
	Path   string         `toml:",omitempty"`
	Regex  *regexp.Regexp `toml:",omitempty"`
}

func (vo *VideoOpt) UnmarshalTOML(decode func(interface{}) error) error {
	type rawVideoOpt struct {
		SiteId string
		Host   string
		Path   string `toml:",omitempty"`
		Regex  string `toml:",omitempty"`
	}

	var err error
	rvo := rawVideoOpt{}
	if err = decode(&rvo); err != nil {
		return err
	}

	var re *regexp.Regexp = nil
	if rvo.Regex != "" {
		re, err = regexp.Compile(rvo.Regex)
		if err != nil {
			return err
		}
	}

	vo.SiteId = rvo.SiteId
	vo.Host = rvo.Host
	vo.Path = rvo.Path
	vo.Regex = re

	if vo.Host == "" {
		return fmt.Errorf("Missing 'host' parameter: %s", vo.SiteId)
	}
	if vo.Path == "" && vo.Regex == nil {
		return fmt.Errorf("Missing 'path' or 'regex' parameter: %s", vo.SiteId)
	}
	return nil
}

type IframeOpt struct {
	SiteId string
	Host   string
	Type   string
	Path   string         `toml:",omitempty"`
	Query  string         `toml:",omitempty"`
	Regex  *regexp.Regexp `toml:",omitempty"`
	Player string
}

func (ifo *IframeOpt) UnmarshalTOML(decode func(interface{}) error) error {
	type rawIframeOpt struct {
		SiteId string
		Host   string
		Type   string
		Path   string `toml:",omitempty"`
		Query  string `toml:",omitempty"`
		Regex  string `toml:",omitempty"`
		Player string
	}

	var err error
	rifo := rawIframeOpt{}
	if err = decode(&rifo); err != nil {
		return err
	}

	var re *regexp.Regexp = nil
	if rifo.Regex != "" {
		re, err = regexp.Compile(rifo.Regex)
		if err != nil {
			return err
		}
	}

	ifo.SiteId = rifo.SiteId
	ifo.Host = rifo.Host
	ifo.Type = rifo.Type
	ifo.Path = rifo.Path
	ifo.Query = rifo.Query
	ifo.Regex = re
	ifo.Player = rifo.Player

	if ifo.Host == "" {
		return fmt.Errorf("Missing 'host' parameter: %s", ifo.SiteId)
	}
	switch ifo.Type {
	case "path":
		if ifo.Path == "" {
			return fmt.Errorf("No found 'path' parameter: %s", ifo.SiteId)
		}
	case "query":
		if ifo.Path == "" {
			return fmt.Errorf("No found 'path' parameter: %s", ifo.SiteId)
		}
		if ifo.Query == "" {
			return fmt.Errorf("No found 'query' parameter: %s", ifo.SiteId)
		}
	case "regex":
		if ifo.Regex == nil {
			return fmt.Errorf("No found 'regex' parameter: %s", ifo.SiteId)
		}
	default:
		return fmt.Errorf("No suppoted 'type' value: %s", ifo.Type)
	}
	return nil
}

type MdConfig struct {
	Extension ExtFlags
	AutoIds   AutoIdsOptions
	Footnote  FootnoteOptions
	Emoji     EmojiOptions
	Embed     EmbedOptions

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
		panic("bad default config file: " + err.Error())
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
