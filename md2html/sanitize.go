package md2html

import (
	_ "embed"

	"github.com/naoina/toml"
	"github.com/sym01/htmlsanitizer"
)

//go:embed htmlsanitizer.conf
var htmlAllowToml []byte

var htmlAllowList = mustUnmarshalAllowList(htmlAllowToml)

func mustUnmarshalAllowList(cnf []byte) *htmlsanitizer.AllowList {
	v := &htmlsanitizer.AllowList{}
	err := toml.Unmarshal(cnf, v)
	if err != nil {
		panic("Unmarshal error: " + err.Error())
	}

	return v
}

type sanitaizer struct {
	impl *htmlsanitizer.HTMLSanitizer
}

func newSanitizer() *sanitaizer {
	impl := htmlsanitizer.NewHTMLSanitizer()
	impl.AllowList = htmlAllowList.Clone()
	return &sanitaizer{impl: impl}
}

func (san *sanitaizer) Sanitize(src_html []byte) ([]byte, error) {
	return san.impl.Sanitize(src_html)
}
