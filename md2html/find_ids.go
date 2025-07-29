package md2html

import (
	"io"
	"strings"

	"golang.org/x/net/html"
)

func FindHtmlIds(r io.Reader) ([]string, error) {
	root, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	fid := newFindIds()
	fid.Find(root)
	return fid.Sum(), nil
}

type findIds struct {
	ids []string
}

func newFindIds() *findIds {
	return &findIds{ ids: []string{} }
}

func (fid *findIds) Sum() []string {
	return fid.ids
}

func (fid *findIds) Find(node *html.Node) {
e_loop:
	for e := node.FirstChild; e != nil; e = e.NextSibling {
		if e.Type != html.ElementNode {
			continue e_loop
		}

		for _, attr := range e.Attr {
			n := strings.ToLower(attr.Key)
			if n == "id" {
				fid.ids = append(fid.ids, attr.Val)
				break
			}
		}
		fid.Find(e)
	}
}
