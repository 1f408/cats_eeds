package ms_include

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func unescape(src []byte) []byte {
	if len(src) == 0 {
		return nil
	}

	dst := make([]byte, len(src))

	w := 0
	for r := 0; r < len(src); r++ {
		if src[r] == '\\' && r+1 < len(src) {
			r++
		}
		dst[w] = src[r]
		w++
	}

	return dst[:w]
}

var includeLabelStateKey = parser.NewContextKey()

type includeLabelState struct {
	ast.BaseInline

	Segment text.Segment
	Target  int

	Prev *includeLabelState

	Next *includeLabelState

	First *includeLabelState

	Last *includeLabelState
}

func newIncludeLabelState(segment text.Segment, tgt int) *includeLabelState {
	return &includeLabelState{
		Segment: segment,
		Target:  tgt,
	}
}

func (s *includeLabelState) Text(source []byte) []byte {
	return s.Segment.Value(source)
}

func (s *includeLabelState) Dump(source []byte, level int) {
	fmt.Printf("%slinkLabelState: \"%s\"\n", strings.Repeat("    ", level), s.Text(source))
}

var kindIncludeLabelState = ast.NewNodeKind("IncludeLabelState")

func (s *includeLabelState) Kind() ast.NodeKind {
	return kindIncludeLabelState
}

func includeLabelStateLength(v *includeLabelState) int {
	if v == nil || v.Last == nil || v.First == nil {
		return 0
	}
	return v.Last.Segment.Stop - v.First.Segment.Start
}

func pushIncludeLabelState(pc parser.Context, v *includeLabelState) {
	labels := pc.Get(includeLabelStateKey)
	var lbl *includeLabelState
	if labels == nil {
		lbl = v
		v.First = v
		v.Last = v
		pc.Set(includeLabelStateKey, lbl)
	} else {
		lbl = labels.(*includeLabelState)
		l := lbl.Last
		lbl.Last = v
		l.Next = v
		v.Prev = l
	}
}

func removeIncludeLabelState(pc parser.Context, d *includeLabelState) {
	labels := pc.Get(includeLabelStateKey)
	if labels == nil {
		return
	}
	lbl := labels.(*includeLabelState)

	if d.Prev == nil {
		lbl = d.Next
		if lbl != nil {
			lbl.First = d
			lbl.Last = d.Last
			lbl.Prev = nil
			pc.Set(includeLabelStateKey, lbl)
		} else {
			pc.Set(includeLabelStateKey, nil)
		}
	} else {
		d.Prev.Next = d.Next
		if d.Next != nil {
			d.Next.Prev = d.Prev
		}
	}
	if lbl != nil && d.Next == nil {
		lbl.Last = d.Prev
	}
	d.Next = nil
	d.Prev = nil
	d.First = nil
	d.Last = nil
}

type msIncludeParser struct {
}

var defaultMsIncludeParser = &msIncludeParser{}

// NewMsIncludeParser return a new InlineParser that parses include.
func NewMsIncludeParser() parser.InlineParser {
	return defaultMsIncludeParser
}

var msIncludeOpenRegexp = regexp.MustCompile(`^\[!INCLUDE\s+\[`)

func indexIncludeOpenEnd(line []byte) int {
	m := msIncludeOpenRegexp.FindIndex(line)
	if m == nil {
		return -1
	}

	return m[1] - 1
}

func (s *msIncludeParser) Trigger() []byte {
	return []byte{'[', ']'}
}

var includeBottom = parser.NewContextKey()

func (s *msIncludeParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, segment := block.PeekLine()

	if line[0] == '[' {
		end := indexIncludeOpenEnd(line)
		if end < 0 {
			return nil
		}

		tgt := indexTitleEnd(line[end:])
		if tgt < 0 {
			return nil
		}

		block.Advance(end)
		pushIncludeBottom(pc)
		return processIncludeLabelOpen(block, segment.Start, segment.Start+end, segment.Start+end+tgt, pc)
	}

	// line[0] == ']'
	labels := pc.Get(includeLabelStateKey)
	if labels == nil {
		return nil
	}

	last := labels.(*includeLabelState).Last
	if last == nil {
		_ = popIncludeBottom(pc)
		return nil
	}

	for {
		if last.Target > segment.Start {
			return nil
		}

		if last.Target == segment.Start {
			break
		}

		_ = popIncludeBottom(pc)
		last = last.Prev

		if last == nil {
			return nil
		}
	}

	block.Advance(1)
	removeIncludeLabelState(pc, last)

	// CommonMark spec says:
	//  > A link label can have at most 999 characters inside the square brackets.
	if includeLabelStateLength(labels.(*includeLabelState)) > 998 {
		ast.MergeOrReplaceTextSegment(last.Parent(), last, last.Segment)
		_ = popIncludeBottom(pc)
		return nil
	}

	c := block.Peek()
	if c != '(' {
		ast.MergeOrReplaceTextSegment(last.Parent(), last, last.Segment)
		_ = popIncludeBottom(pc)
		return nil
	}

	include := s.parseInclude(parent, last, block, pc)
	if include == nil {
		ast.MergeOrReplaceTextSegment(last.Parent(), last, last.Segment)
		_ = popIncludeBottom(pc)
		return nil
	}

	var n ast.Node
	last.Parent().RemoveChild(last.Parent(), last)
	n = include
	n.(interface{ SetPos(int) }).SetPos(last.Segment.Start)
	return n
}

func processIncludeLabelOpen(block text.Reader, start int, end int, tgt int, pc parser.Context) *includeLabelState {
	state := newIncludeLabelState(text.NewSegment(start, end+1), tgt)
	pushIncludeLabelState(pc, state)
	block.Advance(1)
	return state
}

func (s *msIncludeParser) processIncludeLabel(parent ast.Node, link *IncludeNode, last *includeLabelState, pc parser.Context) {
	bottom := popIncludeBottom(pc)
	parser.ProcessDelimiters(bottom, pc)
	for c := last.NextSibling(); c != nil; {
		next := c.NextSibling()
		parent.RemoveChild(parent, c)
		link.AppendChild(link, c)
		c = next
	}
}

var linkFindClosureOptions text.FindClosureOptions = text.FindClosureOptions{
	Nesting: true,
	Newline: false,
	Advance: true,
}

func (s *msIncludeParser) parseInclude(parent ast.Node, last *includeLabelState, block text.Reader, pc parser.Context) *IncludeNode {
	block.Advance(1) // skip '('
	block.SkipSpaces()
	var title []byte
	var destination []byte
	var ok bool
	if block.Peek() == ')' { // empty link like '[link]()'
		block.Advance(1)
	} else {
		destination, ok = parseIncludeDestination(block)
		if !ok {
			return nil
		}
		block.SkipSpaces()
		if block.Peek() != ')' {
			return nil
		}
		block.Advance(1)
	}

	if block.Peek() != ']' {
		return nil
	}

	include := NewIncludeNode()
	s.processIncludeLabel(parent, include, last, pc)
	include.Destination = destination
	include.Title = title
	include.Link = unescape(destination)

	block.Advance(1)
	return include
}

func parseIncludeDestination(block text.Reader) ([]byte, bool) {
	block.SkipSpaces()
	line, _ := block.PeekLine()
	if block.Peek() == '<' {
		i := 1
		for i < len(line) {
			c := line[i]
			if c == '\\' && i < len(line)-1 && util.IsPunct(line[i+1]) {
				i += 2
				continue
			} else if c == '>' {
				block.Advance(i + 1)
				return line[1:i], true
			}
			i++
		}
		return nil, false
	}
	opened := 0
	i := 0
	for i < len(line) {
		c := line[i]
		if c == '\\' && i < len(line)-1 && util.IsPunct(line[i+1]) {
			i += 2
			continue
		} else if c == '(' {
			opened++
		} else if c == ')' {
			opened--
			if opened < 0 {
				break
			}
		} else if util.IsSpace(c) {
			break
		}
		i++
	}
	block.Advance(i)
	return line[:i], len(line[:i]) != 0
}

func pushIncludeBottom(pc parser.Context) {
	bottoms := pc.Get(includeBottom)
	b := pc.LastDelimiter()
	if bottoms == nil {
		pc.Set(includeBottom, b)
		return
	}
	if s, ok := bottoms.([]ast.Node); ok {
		pc.Set(includeBottom, append(s, b))
		return
	}
	pc.Set(includeBottom, []ast.Node{bottoms.(ast.Node), b})
}

func popIncludeBottom(pc parser.Context) ast.Node {
	bottoms := pc.Get(includeBottom)
	if bottoms == nil {
		return nil
	}
	if v, ok := bottoms.(ast.Node); ok {
		pc.Set(includeBottom, nil)
		return v
	}
	s := bottoms.([]ast.Node)
	v := s[len(s)-1]
	n := s[0 : len(s)-1]
	switch len(n) {
	case 0:
		pc.Set(includeBottom, nil)
	case 1:
		pc.Set(includeBottom, n[0])
	default:
		pc.Set(includeBottom, s[0:len(s)-1])
	}
	return v
}

func (s *msIncludeParser) CloseBlock(parent ast.Node, block text.Reader, pc parser.Context) {
	pc.Set(includeBottom, nil)
	labels := pc.Get(includeLabelStateKey)
	if labels == nil {
		return
	}
	for s := labels.(*includeLabelState); s != nil; {
		next := s.Next
		removeIncludeLabelState(pc, s)
		s.Parent().ReplaceChild(s.Parent(), s, ast.NewTextSegment(s.Segment))
		s = next
	}
}

func indexTitleEnd(line []byte) int {
	if len(line) < 2 {
		return -1
	}
	if line[0] != '[' {
		return -2
	}

	tgt := "[]\\"

	cur := 1
	dep := 0
search:
	for {
		pos, nc, size := findAnyASCII(line[cur:], tgt)
		if pos < 0 {
			return -3
		}
		cur += pos

		switch line[cur] {
		case '[':
			dep++
			cur++
		case ']':
			if dep <= 0 {
				if nc != '(' {
					return -4
				}
				break search
			}
			dep--
			cur++
		case '\\':
			cur += size
		}
	}

	if len(line[cur:]) < 1 {
		return -5
	}

	return cur
}

func findAnyASCII(b []byte, tgt string) (int, rune, int) {
	pos := bytes.IndexAny(b, tgt)
	if pos == -1 {
		return -1, utf8.RuneError, 0
	}

	var nextRune rune
	var size int
	if pos+1 < len(b) {
		nextRune, size = utf8.DecodeRune(b[pos+1:])
	}

	return pos, nextRune, size
}
