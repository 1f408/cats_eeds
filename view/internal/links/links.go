package links

import (
	"path"
	"strings"

	"github.com/l4go/rpath"
)

type Link struct {
	Name string
	Path string
}

func NewLinks(p string) []Link {
	p = rpath.Clean(p)
	if p[0] != '/' {
		return nil
	}

	is_dir := rpath.IsDir(p)

	dirs := []string{"/"}
	if p != "/" {
		dirs = strings.Split(path.Clean(p), "/")
		dirs[0] = "/"
	}

	if is_dir {
		return new_links_by_dir(dirs)
	}
	return new_links_by_file(dirs)
}

func new_links_by_file(dirs []string) []Link {
	links := make([]Link, len(dirs))
	for i, n := range dirs {
		rev_i := len(dirs) - i - 1

		links[i].Name = n
		switch rev_i {
		case 0:
			links[i].Path = ""
		case 1:
			links[i].Path = "."
		default:
			links[i].Path = encode_file_url(strings.Repeat("../", rev_i-2) + "..")
		}
	}

	return links
}

func new_links_by_dir(dirs []string) []Link {
	links := make([]Link, len(dirs))
	for i, n := range dirs {
		rev_i := len(dirs) - i - 1

		links[i].Name = n
		switch rev_i {
		case 0:
			links[i].Path = ""
		default:
			links[i].Path = encode_file_url(strings.Repeat("../", rev_i-1) + "..")
		}
	}

	return links
}

func is_encode_file_url(c byte) bool {
	if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' ||
		c == '-' || c == '.' || c == '_' || c == '~' || c == '/' {
		return false
	}

	return true
}

func EncodeFileURL(p string) string {
	p = rpath.Clean(p)

	if p == "" {
		return p
	}
	if p[0] != '/' {
		return "./" + encode_file_url(p)
	}

	return encode_file_url(p)
}

func encode_file_url(src string) string {
	const hex_char = "0123456789ABCDEF"

	esc_cnt := 0

	for i := 0; i < len(src); i++ {
		c := src[i]
		if is_encode_file_url(c) {
			esc_cnt++
		}
	}
	if esc_cnt == 0 {
		return src
	}

	req_len := len(src) + 2*esc_cnt
	dst := make([]byte, req_len)

	j := 0
	for i := 0; i < len(src); i++ {
		c := src[i]
		if is_encode_file_url(c) {
			dst[j] = '%'
			dst[j+1] = hex_char[c>>4]
			dst[j+2] = hex_char[c&0xF]
			j += 3
		} else {
			dst[j] = c
			j++
		}
	}

	return string(dst)
}
