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
			links[i].Path = strings.Repeat("../", rev_i-2) + ".."
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
			links[i].Path = strings.Repeat("../", rev_i-1) + ".."
		}
	}

	return links
}
