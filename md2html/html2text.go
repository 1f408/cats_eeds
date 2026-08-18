package md2html

import (
	"io"

	"golang.org/x/net/html"
)

func convertHtmlNodeText(e *html.Node, w io.Writer) error {
	if e.Type == html.TextNode {
		_, err := io.WriteString(w, e.Data)
		return err
	}

	for c := e.FirstChild; c != nil; c = c.NextSibling {
		err := convertHtmlNodeText(c, w)
		if err != nil {
			return err
		}
	}
	return nil
}
