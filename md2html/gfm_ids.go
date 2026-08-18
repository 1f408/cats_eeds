package md2html

import (
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"golang.org/x/text/unicode/norm"

	"github.com/1f408/cats_eeds/md2html/uniqid"
)

type GfmIDs struct {
	parser goldmark.Markdown
	tbl uniqid.IdsTable
}

func init() {
	AutoIdsMap["gfm"] = NewGfmIDs
}

func NewGfmIDs(p goldmark.Markdown, sys_id_tbl uniqid.IdsTable) parser.IDs {
	return &GfmIDs{
		parser: p,
		tbl: sys_id_tbl,
	}
}

func (ids *GfmIDs) toText(value []byte) []byte {
	return toPrintableBytes(value)
}

func (ids *GfmIDs) toValid(value []byte) []byte {
	ancher := make([]byte, 0, len(value))

	dash_mode := false
	for _, r := range string(value) {
		if r == '-' || r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r) {
			if dash_mode && len(ancher) > 0 {
				ancher = append(ancher, '-')
			}

			dash_mode = false
			ancher = append(ancher, string(unicode.ToLower(r))...)
		} else if r != ' ' {
			if dash_mode && len(ancher) > 0 {
				ancher = append(ancher, '-')
			}

			dash_mode = false
		} else {
			dash_mode = true
		}
	}
	return ancher
}

func (ids *GfmIDs) Generate(value []byte, kind ast.NodeKind) []byte {
	value = ids.toText(value)
	value = ids.toValid(value)
	value = norm.NFC.Bytes(value)

	return uniqid.Generate(ids.tbl, value)
}

func (ids *GfmIDs) Has(value []byte) bool {
	return ids.tbl.Has(value)
}

func (ids *GfmIDs) Put(value []byte) {
	ids.tbl.Put(value)
}
