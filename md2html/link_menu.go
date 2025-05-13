package md2html

import (
	"github.com/goccy/go-yaml"

	"github.com/1f408/cats_eeds/internal/perenc"
)

type LinkMenuConfig struct {
	Default []Link `toml:",omitempty"`
}

type Link struct {
	Label     string `yaml:"label,omitempty" toml:"label,omitempty" json:"label,omitempty"`
	Url       string `yaml:"url,omitempty" toml:"url,omitempty" json:"url,omitempty"`
}

func (l *Link) fix() {
	if l == nil {
		return
	}

	l.Url = perenc.EncodeUrl(l.Url)
}

func (l *Link) UnmarshalTOML(decode func(interface{}) error) error {
	type rawLink Link
	if err := decode((*rawLink)(l)); err != nil {
		return err
	}

	l.fix()
	return nil
}

func (l *Link) UnmarshalYAML(src []byte) error {
	type rawLink Link
	v := Link{}
	if e := yaml.Unmarshal(src, (*rawLink)(&v)); e != nil {
		return e
	}

	v.fix()
	*l = v
	return nil
}
