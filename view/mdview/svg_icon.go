package mdview

import (
	"regexp"
	"strings"
	"unicode"
)

var tmpl_svg_icon_type_reg = regexp.MustCompile(`^[a-z0-9_\-]+$`)
var tmpl_svg_icon_cache = map[string]string{}

func trim_right_sp(str string) string {
	return strings.TrimRightFunc(str, unicode.IsSpace)
}

func (mdv *MdView) TmplSvgIcon(name string) string {
	if mdv.SvgIconPath.IsZero() {
		return ""
	}

	if !tmpl_svg_icon_type_reg.MatchString(name) {
		return ""
	}

	if svg, ok := tmpl_svg_icon_cache[name]; ok {
		return svg
	}

	var svg string = ""
	if file, jerr := mdv.SvgIconPath.Join(name + ".svg"); jerr == nil {
		if bin, rerr := file.ReadFile(mdv.SystemFS); rerr == nil {
			svg = trim_right_sp(string(bin))
		}
	}

	tmpl_svg_icon_cache[name] = svg
	return svg
}
