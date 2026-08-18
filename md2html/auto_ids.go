package md2html

import (
	"errors"
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"

	"github.com/1f408/cats_eeds/md2html/uniqid"
)

var ErrBadAutoIdsType = errors.New("bad auto IDs type")

var AutoIdsMap = map[string]func(parser goldmark.Markdown, sys_ids_tbl uniqid.IdsTable) parser.IDs{}

func NewAutoIds(parser goldmark.Markdown, id_type string, sys_id_tbl uniqid.IdsTable) (parser.IDs, error) {
	new_func, ok := AutoIdsMap[id_type]
	if !ok {
		return nil, ErrBadAutoIdsType
	}

	return new_func(parser, sys_id_tbl), nil
}

func toPrintableString(value string) string {
	is_good := func(r rune) rune {
		if !unicode.IsPrint(r) {
			return -1
		}
		return r
	}

	return strings.Map(is_good, value)
}

func toPrintableBytes(value []byte) []byte {
	return []byte(toPrintableString(string(value)))
}
