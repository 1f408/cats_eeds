package md2html

import (
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"golang.org/x/text/unicode/norm"

	"github.com/1f408/cats_eeds/md2html/uniqid"
)

type SafeIDs struct {
	parser goldmark.Markdown
	tbl uniqid.IdsTable
}

func init() {
	AutoIdsMap[""] = NewSafeIDs
	AutoIdsMap["safe"] = NewSafeIDs
}

func NewSafeIDs(p goldmark.Markdown, sys_id_tbl uniqid.IdsTable) parser.IDs {
	return &GfmIDs{
		parser: p,
		tbl: sys_id_tbl,
	}
}

func (ids *SafeIDs) toText(value []byte) []byte {
	return toPrintableBytes(value)
}

func (ids *SafeIDs) toValid(value []byte) []byte {
	anchor := make([]byte, 0, len(value))

	for _, r := range string(value) {
		if unicode.IsPrint(r) {
			anchor = append(anchor, string(r)...)
		}
	}
	return anchor
}

func (ids *SafeIDs) Generate(value []byte, kind ast.NodeKind) []byte {
	value = ids.toText(value)
	value = ids.toValid(value)
	value = norm.NFC.Bytes(value)

	return uniqid.Generate(ids.tbl, value)
}

func (ids *SafeIDs) Has(value []byte) bool {
	return ids.tbl.Has(value)
}

func (ids *SafeIDs) Put(value []byte) {
	ids.tbl.Put(value)
}
