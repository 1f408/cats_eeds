package md2html

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/naoina/toml"

	"github.com/1f408/cats_eeds/frontmatter"
)

var ErrInvalidFrontMatter = errors.New("invalid front matter")

type UnmarshalFunc func([]byte, any) error

func tomlUnmarshal(data []byte, v interface{}) error {
	cfg := toml.DefaultConfig
	cfg.MissingField = func(_ reflect.Type, _ string) error { return nil }
	return cfg.Unmarshal(data, v)
}

var metaFormats = map[string]UnmarshalFunc{
	"---": yaml.Unmarshal,
	"+++": tomlUnmarshal,
	";;;": json.Unmarshal,
}

const CatsSeriesName = "cats"

type CustomParamConfig struct {
	Default CustomParam `toml:",omitempty"`
}
type CustomParam map[string]any

type FrontMatterParam struct {
	Product           string  `yaml:"product,omitempty" toml:"product,omitempty" json:"product,omitempty"`
	Title             string  `yaml:"title,omitempty" toml:"title,omitempty" json:"title,omitempty"`
	MarkdownConfig    string  `yaml:"markdown_config,omitempty" toml:"markdown_config,omitempty" json:"markdown_config,omitempty"`
	ThemeStyle        string  `yaml:"theme_style,omitempty" toml:"theme_style,omitempty" json:"theme_style,omitempty"`
	LocationNavi      string  `yaml:"location_navi,omitempty" toml:"location_navi,omitempty" json:"location_navi,omitempty"`
	TocNavi           string  `yaml:"toc_navi,omitempty" toml:"toc_navi,omitempty" json:"toc_navi,omitempty"`
	PageStyle         string  `yaml:"page_style,omitempty" toml:"page_style,omitempty" json:"page_style,omitempty"`
	DirectoryViewMode string  `yaml:"directory_view_mode,omitempty" toml:"directory_view_mode,omitempty" json:"directory_view_mode,omitempty"`
	PaperType         string  `yaml:"paper_type,omitempty" toml:"paper_type,omitempty" json:"paper_type,omitempty"`
	PrintZoom         float32 `yaml:"print_zoom,omitempty" toml:"print_zoom,omitempty" json:"print_zoom,omitempty"`

	SmCard      SmCardParam `yaml:"sm_card,omitempty" toml:"sm_card,omitempty" json:"sm_card,omitempty"`
	CustomParam CustomParam `yaml:"custom_param,omitempty" toml:"custom_param,omitempty" json:"custom_param,omitempty"`
	LinkMenu    []Link      `yaml:"link_menu,omitempty" toml:"link_menu,omitempty" json:"link_menu,omitempty"`

	Config *FrontMatterConfig `yaml:"-" toml:"-" json:"-"`
}

type FrontMatterConfig struct {
	Yaml bool `toml:",omitempty"`
	Toml bool `toml:",omitempty"`
	Json bool `toml:",omitempty"`

	UsedForHtml bool `toml:",omitempty"`
	UsedForText bool `toml:",omitempty"`
}

var zeroFrontMatterConfig = FrontMatterConfig{}

func (fmc *FrontMatterConfig) IsEnabled() bool {
	return fmc != nil && *fmc != zeroFrontMatterConfig
}

func (fmc *FrontMatterConfig) delimers() []string {
	fm_words := make([]string, 0, 3)
	if fmc.Yaml {
		fm_words = append(fm_words, "---")
	}
	if fmc.Toml {
		fm_words = append(fm_words, "+++")
	}
	if fmc.Json {
		fm_words = append(fm_words, ";;;")
	}

	return fm_words
}

func (fmc *FrontMatterConfig) TrimAndParse(bin []byte) ([]byte, *FrontMatterParam, error) {
	r := bytes.NewReader(bin)
	fm := frontmatter.New(fmc.delimers())
	wd, head, body, err := fm.FindAndSplit(r)
	if err != nil {
		return bin, nil, err
	}

	fn, ok := metaFormats[wd]
	if !ok {
		return bin, nil, ErrInvalidFrontMatter
	}

	fmp := &FrontMatterParam{}
	if perr := fn(head, fmp); err != nil {
		return bin, nil, perr
	}

	if !fmp.validate() {
		fmp = &FrontMatterParam{}
	}

	fmp.Config = fmc
	fmp.fix()

	return body, fmp, nil
}

func (fmp *FrontMatterParam) validate() bool {
	if fmp.Product != CatsSeriesName {
		return false
	}

	return true
}

func (fmp *FrontMatterParam) fix() {
	switch fmp.ThemeStyle {
	case "radio":
	case "load":
	case "os":
	default:
		fmp.ThemeStyle = ""
	}

	switch fmp.DirectoryViewMode {
	case "none":
	case "autoindex":
	case "close":
	case "auto":
	case "open":
	default:
		fmp.DirectoryViewMode = ""
	}

	switch fmp.LocationNavi {
	case "none":
	case "dirs":
	default:
		fmp.LocationNavi = ""
	}

	switch fmp.TocNavi {
	case "none":
	case "details":
	case "static":
	default:
		fmp.TocNavi = ""
	}

	if strings.ContainsRune(fmp.PageStyle, '/') {
		fmp.PageStyle = ""
	}
}
