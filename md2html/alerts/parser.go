package alerts

import (
	"regexp"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type alertBlockParser struct {
	Config
}

func NewAlertBlockParser(opts ...Option) parser.BlockParser {
	p := &alertBlockParser{
		Config: DefaultConfig,
	}
	for _, opt := range opts {
		opt.SetAlertBlockOption(&p.Config)
	}
	return p
}

func (b *alertBlockParser) Trigger() []byte {
	return []byte{'>'}
}

var alertBlockRegexp = regexp.MustCompile(`^>\s*\[!(\w+)\]\s*$`)

func (b *alertBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	w, pos := util.IndentWidth(line, reader.LineOffset())
	if w > 3 || pos >= len(line) || line[pos] != '>' {
		return nil, parser.NoChildren
	}

	m := alertBlockRegexp.FindSubmatchIndex(line[pos:])
	if m == nil {
		return nil, parser.NoChildren
	}

	ind, _ := util.IndentWidth(line[pos+1:], reader.LineOffset()+pos+1)
	if ind > 4 {
		return nil, parser.NoChildren
	}

	lbl := line[m[2]:m[3]]
	if _, ok := b.Config.TitleHtmlMapping[string(lbl)]; !ok {
		return nil, parser.NoChildren
	}

	an := NewAlertBlockNode(lbl)
	an.AppendChild(an, NewAlertTitleNode(lbl))

	reader.Advance(pos + 1)
	return an, parser.NoChildren
}

func (b *alertBlockParser) read_quote(reader text.Reader) bool {
	line, _ := reader.PeekLine()
	w, pos := util.IndentWidth(line, reader.LineOffset())
	if w > 3 || pos >= len(line) || line[pos] != '>' {
		return false
	}

	pos++
	if pos >= len(line) || line[pos] == '\n' {
		reader.Advance(pos)
		return true
	}
	reader.Advance(pos)
	if line[pos] == ' ' || line[pos] == '\t' {
		padding := 0
		if line[pos] == '\t' {
			padding = util.TabWidth(reader.LineOffset()) - 1
		}
		reader.AdvanceAndSetPadding(1, padding)
	}
	return true
}

func (b *alertBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	if b.read_quote(reader) {
		return parser.Continue | parser.HasChildren
	}
	return parser.Close
}

func (b *alertBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
}

func (b *alertBlockParser) CanInterruptParagraph() bool {
	return true
}

func (b *alertBlockParser) CanAcceptIndentedLine() bool {
	return false
}
