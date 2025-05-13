package tmplext

import (
	"io/fs"
	"regexp"
	"strings"
	"unicode"

	"github.com/1f408/cats_eeds/upath"
)

var tmpl_svg_icon_type_reg = regexp.MustCompile(`^[a-z0-9_\-]+$`)
var tmpl_svg_icon_cache = map[string]string{}

func trim_right_sp(str string) string {
	return strings.TrimRightFunc(str, unicode.IsSpace)
}

func TmplDummySvgIcon(_ string) string {
	return ""
}

func NewTmplSvgIcon(fsys fs.FS, up upath.UPath) func(name string) string {
	fn := func(name string) string {
		if up.IsZero() {
			return ""
		}

		if !tmpl_svg_icon_type_reg.MatchString(name) {
			return ""
		}

		if svg, ok := tmpl_svg_icon_cache[name]; ok {
			return svg
		}

		var svg string = ""
		if file, jerr := up.Join(name + ".svg"); jerr == nil {
			if bin, rerr := file.ReadFile(fsys); rerr == nil {
				svg = trim_right_sp(string(bin))
			}
		}

		tmpl_svg_icon_cache[name] = svg
		return svg
	}

	return fn
}
