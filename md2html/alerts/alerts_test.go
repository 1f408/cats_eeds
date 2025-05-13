package alerts

import (
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/testutil"
)

func TestAlertBlock(t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			NewAlertBlock(),
		),
	)
	testutil.DoTestCaseFile(markdown, "_test/alert_block.txt", t, testutil.ParseCliCaseArg()...)
}
