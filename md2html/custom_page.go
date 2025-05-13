package md2html

import (
	_ "embed"

	"github.com/naoina/toml"
)

//go:embed custom_page.conf
var defaultCustomPage []byte

type CustomPageConfig struct {
	FrontMatter FrontMatterConfig `toml:",omitempty"`
	SmCard      SmCardConfig      `toml:",omitempty"`
	LinkMenu    LinkMenuConfig    `toml:",omitempty"`
	PrintPaper  PrintPaperConfig  `toml:",omitempty"`
	CustomParam CustomParamConfig `toml:",omitempty"`

	PageStyle string  `toml:",omitempty"`

	init bool `toml:"-"`
}

func (cpc *CustomPageConfig) Initialize() {
	type Raw CustomPageConfig
	err := toml.Unmarshal(defaultCustomPage, (*Raw)(cpc))
	if err != nil {
		panic("bad default config file: " + err.Error())
	}

	cpc.init = true
}

func (_ *CustomPageConfig) MakeNew() *CustomPageConfig {
	fmc := &CustomPageConfig{}
	fmc.Initialize()

	return fmc
}

func (cpc *CustomPageConfig) UnmarshalTOML(decode func(interface{}) error) error {
	if !cpc.init {
		cpc.Initialize()
	}

	type Raw CustomPageConfig
	return decode((*Raw)(cpc))
}
