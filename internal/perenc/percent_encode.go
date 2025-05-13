package perenc

import (
	"bytes"
	"net/url"
	"regexp"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/l4go/rpath"
)

var escapeDataUriMap = map[rune]bool{
	' ': true, '"': true, '<': true, '>': true, // for HTML
	':': true, '/': true, '?': true, '#': true,
	'[': true, ']': true, '@': true,
	'!': true, '$': true, '&': true, '\'': true, '(': true, ')': true,
	'*': true, '+': true, ',': true, ';': true, '=': true,
}

func is_escape_uri_data(r rune) bool {
	return 0 <= r && r <= 0x1F || r == 0x7F || escapeDataUriMap[r]
}

func EncodeUriData(src string) string {
	return EncodeRune(is_escape_uri_data, src)
}

var escapeUrlFragmentMap = map[rune]bool{
	// "fragment percent-encode set" defined in URL Living Standard
	' ': true, '"': true, '<': true, '>': true, '`': true,
	// for HTML
	'&': true, '\'': true,
}

func is_escape_url_fragment(r rune) bool {
	return 0 <= r && r <= 0x1F || r > 0x7E || escapeUrlFragmentMap[r]
}

func EncodeUrlFragment(src string) string {
	if src == "" {
		return src
	}
	return EncodeRune(is_escape_url_fragment, src)
}

var escapeUrlPathMap = map[rune]bool{
	// "path percent-encode set" defined in URL Living Standard
	' ': true, '"': true, '#': true, '<': true, '>': true,
	'?': true, '`': true, '{': true, '}': true,
	// for HTML
	'&': true, '\'': true,
}

func is_escape_url_path(r rune) bool {
	return 0 <= r && r <= 0x1F || r > 0x7E || escapeUrlPathMap[r]
}

var reScheme = regexp.MustCompile(`^[a-z0-9]+://`)

func EncodeUrlPath(src string) string {
	if src == "" {
		return src
	}
	if !reScheme.MatchString(src) {
		src = rpath.Clean(src)
		switch {
		case src[0] == '/':
		case src == ".":
		case src == "..":
		case strings.HasPrefix(src, "./"):
		case strings.HasPrefix(src, "../"):
		default:
			src = "./" + src
		}
	}

	return EncodeRune(is_escape_url_path, src)
}

const hex_char = "0123456789ABCDEF"

func escape_rune(r rune) []byte {

	src := make([]byte, 0, utf8.UTFMax)
	src = utf8.AppendRune(src, r)

	var dst bytes.Buffer
	for _, c := range src {
		dst.WriteByte('%')
		dst.WriteByte(hex_char[c>>4])
		dst.WriteByte(hex_char[c&0xF])
	}

	return dst.Bytes()
}

func EncodeRune(is_esc func(rune) bool, src string) string {
	has_esc := false
	for _, r := range src {
		if is_esc(r) {
			has_esc = true
		}
	}
	if !has_esc {
		return src
	}

	var b strings.Builder
	for _, r := range src {
		if r == '%' || is_esc(r) {
			b.Write(escape_rune(r))
		} else {
			b.WriteRune(r)
		}
	}

	return b.String()
}

func EncodeUrl(str string) string {
	u, e := url.Parse(str)
	if e != nil {
		return ""
	}

	return u.String()
}

func EncodeHref(str string) string {
	u, e := url.Parse(str)
	if e != nil {
		return ""
	}

	return template.HTMLEscapeString(u.String())
}
