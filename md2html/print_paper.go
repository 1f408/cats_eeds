package md2html

import (
	_ "embed"

	"github.com/naoina/toml"

	"github.com/1f408/cats_eeds/upath"
)

//go:embed "print_paper_mapping.conf"
var defaultPaperMapping []byte

type PrintPaperConfig struct {
	Mapping upath.Import[*PrintPaperMapping] `toml:",omitempty"`
	Default struct {
		PaperType string  `toml:",omitempty"`
		PrintZoom float32 `toml:",omitempty"`
	} `toml:",omitempty"`
}

type PrintPaperMapping map[string]string

func (ppm *PrintPaperMapping) Initialize() {
	type Raw PrintPaperMapping
	if err := toml.Unmarshal(defaultPaperMapping, (*Raw)(ppm)); err != nil {
		panic("bad default paper_size config file: " + err.Error())
	}
}

func (_ *PrintPaperMapping) MakeNew() *PrintPaperMapping {
	ppm := &PrintPaperMapping{}
	ppm.Initialize()

	return ppm
}

func (ppm *PrintPaperMapping) UnmarshalTOML(decode func(interface{}) error) error {
	if ppm == nil {
		ppm.Initialize()
	}

	type Raw PrintPaperMapping
	return decode((*Raw)(ppm))
}

func (ppm *PrintPaperMapping) GetCss(name string) (string, bool) {
	if ppm == nil {
		return "", false
	}

	sz, ok := (*ppm)[name]
	return sz, ok
}
