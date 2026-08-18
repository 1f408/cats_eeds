package uniqid

import (
	"fmt"
	"strings"
)

type IdsTable interface {
	Has(id []byte) bool
	Put(id []byte)
	IsValidPrefix(pfx []byte) bool
	ReservePrefix(pfx []byte)
	CleanPrefix(pfx []byte)
}

type mapIdsTable struct {
	pfx map[string]struct{}
	ids map[string]struct{}
}

func NewMapIdsTable() IdsTable {
	tbl := &mapIdsTable{
		pfx: map[string]struct{}{},
		ids: map[string]struct{}{},
	}

	return tbl
}

func (tbl *mapIdsTable) Has(id []byte) bool {
	if _, pok := tbl.ids[string(id)]; pok {
		return true
	}

	for pf, _ := range tbl.pfx {
		if strings.HasPrefix(string(id), pf) {
			return true
		}
	}

	return false
}

func (tbl *mapIdsTable) Put(id []byte){
	tbl.ids[string(id)] = struct{}{}
}

func (tbl *mapIdsTable) IsValidPrefix(prefix_bin []byte) bool {
	prefix := string(prefix_bin)

	if _, pok := tbl.pfx[prefix]; pok {
		return false
	}

	for pf, _ := range tbl.pfx {
		if strings.HasPrefix(prefix, pf) {
			return false
		}
		if strings.HasPrefix(pf, prefix) {
			return false
		}
	}

	for id, _ := range tbl.ids {
		if strings.HasPrefix(id, prefix) {
			return false
		}
	}

	return true
}

func (tbl *mapIdsTable) ReservePrefix(prefix []byte){
	tbl.pfx[string(prefix)] = struct{}{}
}

func (tbl *mapIdsTable) CleanPrefix(prefix []byte){
	delete(tbl.pfx, string(prefix))
}

func Generate(ids IdsTable, value []byte) []byte {
	if len(value) == 0 {
		value = []byte("id")
	}

	if !ids.Has(value) {
		ids.Put(value)
		return value
	}

	buf := make([]byte, 0, len(value)+3)
	buf = append(buf, value...)
rewrite:
	for i := 1; ; i++ {
		buf = buf[:len(value)]
		buf = fmt.Appendf(buf, "-%d", i)
		if !ids.Has(buf) {
			break rewrite
		}
	}

	ids.Put(buf)
	return buf
}
