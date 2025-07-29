package md2html

import (
	"errors"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
)

var ErrBadAutoIdsType = errors.New("bad auto IDs type")

var AutoIdsMap = map[string]func(parser goldmark.Markdown, sys_ids []string) parser.IDs{}

func NewAutoIds(parser goldmark.Markdown, id_type string, sys_ids []string) (parser.IDs, error) {
	new_func, ok := AutoIdsMap[id_type]
	if !ok {
		return nil, ErrBadAutoIdsType
	}

	return new_func(parser, sys_ids), nil
}
