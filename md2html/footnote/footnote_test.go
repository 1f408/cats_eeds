package footnote_test

import (
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/testutil"

	"github.com/1f408/cats_eeds/md2html/footnote"
	"github.com/1f408/cats_eeds/md2html/uniqid"
)

func TestFootnote(t *testing.T) {
	for _, no := range testutil.ParseCliCaseArg() {
		dmy_id_tbl := uniqid.NewMapIdsTable()
		markdown := goldmark.New(
			goldmark.WithRendererOptions(
				html.WithUnsafe(),
			),
			goldmark.WithExtensions(
				footnote.NewFootnote(dmy_id_tbl),
			),
		)
		testutil.DoTestCaseFile(markdown, "_test/footnote.txt", t, no)
	}
}

func TestFootnoteOptions(t *testing.T) {
	dmy_id_tbl := uniqid.NewMapIdsTable()
	markdown := goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			footnote.NewFootnote(
				dmy_id_tbl,
				footnote.WithFootnoteIDPrefix("my-fn"),
				footnote.WithFootnoteLinkClass("link-class"),
				footnote.WithFootnoteBacklinkClass("backlink-class"),
				footnote.WithFootnoteLinkTitle("link-title-%%-^^"),
				footnote.WithFootnoteBacklinkTitle("backlink-title"),
				footnote.WithFootnoteBacklinkHTML("^"),
			),
		),
	)

	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          1,
			Description: "Footnote with options",
			Markdown: `That's some text with a footnote.[^1]

Same footnote.[^1]

Another one.[^2]

[^1]: And that's the footnote.
[^2]: Another footnote.
`,
Expected: `<p>That's some text with a footnote.<sup id="my-fn:1-r"><a href="#my-fn:1" class="link-class" title="link-title-2-1" role="doc-noteref">1</a></sup></p>
<p>Same footnote.<sup id="my-fn:1-r-1"><a href="#my-fn:1" class="link-class" title="link-title-2-1" role="doc-noteref">1</a></sup></p>
<p>Another one.<sup id="my-fn:2-r"><a href="#my-fn:2" class="link-class" title="link-title-1-2" role="doc-noteref">2</a></sup></p>
<div class="footnotes" role="doc-endnotes">
<hr>
<ol>
<li id="my-fn:1">
<p>And that's the footnote.&#160;<a href="#my-fn:1-r" class="backlink-class" title="backlink-title" role="doc-backlink">^</a>&#160;<a href="#my-fn:1-r-1" class="backlink-class" title="backlink-title" role="doc-backlink">^</a></p>
</li>
<li id="my-fn:2">
<p>Another footnote.&#160;<a href="#my-fn:2-r" class="backlink-class" title="backlink-title" role="doc-backlink">^</a></p>
</li>
</ol>
</div>`,
		},
		t,
	)

	dmy_id_tbl = uniqid.NewMapIdsTable()
	dmy_id_tbl.Put([]byte("fn:1"))
	dmy_id_tbl.ReservePrefix([]byte("fn-1:"))
	markdown = goldmark.New(
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
		goldmark.WithExtensions(
			footnote.NewFootnote(
				dmy_id_tbl,
				footnote.WithFootnoteIDPrefix("fn"),
				footnote.WithFootnoteLinkClass("link-class"),
				footnote.WithFootnoteBacklinkClass("backlink-class"),
				footnote.WithFootnoteLinkTitle("link-title-%%-^^"),
				footnote.WithFootnoteBacklinkTitle("backlink-title"),
				footnote.WithFootnoteBacklinkHTML("^"),
			),
		),
	)

	testutil.DoTestCase(
		markdown,
		testutil.MarkdownTestCase{
			No:          2,
			Description: "Footnote with dup prefix",
			Markdown: `That's some text with a footnote.[^1]

Same footnote.[^1]

Another one.[^2]

[^1]: And that's the footnote.
[^2]: Another footnote.
`,
Expected: `<p>That's some text with a footnote.<sup id="fn-2:1-r"><a href="#fn-2:1" class="link-class" title="link-title-2-1" role="doc-noteref">1</a></sup></p>
<p>Same footnote.<sup id="fn-2:1-r-1"><a href="#fn-2:1" class="link-class" title="link-title-2-1" role="doc-noteref">1</a></sup></p>
<p>Another one.<sup id="fn-2:2-r"><a href="#fn-2:2" class="link-class" title="link-title-1-2" role="doc-noteref">2</a></sup></p>
<div class="footnotes" role="doc-endnotes">
<hr>
<ol>
<li id="fn-2:1">
<p>And that's the footnote.&#160;<a href="#fn-2:1-r" class="backlink-class" title="backlink-title" role="doc-backlink">^</a>&#160;<a href="#fn-2:1-r-1" class="backlink-class" title="backlink-title" role="doc-backlink">^</a></p>
</li>
<li id="fn-2:2">
<p>Another footnote.&#160;<a href="#fn-2:2-r" class="backlink-class" title="backlink-title" role="doc-backlink">^</a></p>
</li>
</ol>
</div>`,
		},
		t,
	)
}
