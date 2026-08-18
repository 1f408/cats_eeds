package ms_include

import (
	"testing"

	"io"
	"errors"
	"path"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/testutil"
)

var ErrDummy = errors.New("dummy error")

func test_convert(file string, w io.Writer) error {
	if path.Base(file) == "loop.md" {
		return ErrDummy
	}
	// dummy convert
	if _, e := io.WriteString(w, "<code>include: "); e != nil {
		return e
	}
	if _, e := io.WriteString(w, file); e != nil {
		return e
	}
	if _, e := io.WriteString(w, "</code>"); e != nil {
		return e
	}

	return nil
}


func TestInclude(t *testing.T) {
	ps := NewSlicePathStack("/var/www/html/README.md")
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			NewMsInclude(test_convert, ps),
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/include_line.txt", t, testutil.ParseCliCaseArg()...)
}
