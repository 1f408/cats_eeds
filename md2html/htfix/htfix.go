package htfix

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

var ErrInvalidHTML = errors.New("invalid html")

func NewFixHTML(r io.Reader, w io.Writer) error {
	doc, err := html.Parse(r)
	if err != nil {
		return err
	}

	var root *html.Node = nil
	for c := doc.FirstChild; c != nil; c = c.NextSibling {
		if c.DataAtom == atom.Html {
			root = c
			break
		}
	}
	if root == nil {
		return ErrInvalidHTML
	}

	var body *html.Node = nil
	for c := root.FirstChild; c != nil; c = c.NextSibling {
		if c.DataAtom == atom.Body {
			body = c
			break
		}
	}
	if body == nil {
		return ErrInvalidHTML
	}

	for c := body.FirstChild; c != nil; c = c.NextSibling {
		err := html.Render(w, c)
		if err != nil {
			return err
		}
	}

	return nil
}

func FixHTML(src []byte) []byte {
	var b bytes.Buffer

	r := bytes.NewReader(src)
	err := NewFixHTML(r, &b)
	if err != nil {
		return nil
	}

	return b.Bytes()
}

func FixHTMLString(src string) string {
	var b strings.Builder

	r := strings.NewReader(src)
	err := NewFixHTML(r, &b)
	if err != nil {
		return ""
	}

	return b.String()
}
