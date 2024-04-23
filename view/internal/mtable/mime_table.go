package mtable

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"mime"
	"unicode"
)

func inNotPrint(text []byte) bool {
	return bytes.IndexFunc(text, func(r rune) bool {
		return !unicode.IsPrint(r)
	}) > 0
}

type MimeExtTable map[string]string

//go:embed "mime_extension_table.conf"
var defaultMimeExtTableBin []byte

var DefaultMimeExtTable = mustNewDefault()

func mustNewDefault() *MimeExtTable {
	v := &MimeExtTable{}
	if err := v.UnmarshalBinary(defaultMimeExtTableBin); err != nil {
		panic("bad default config: " + err.Error())
	}

	return v
}

func (_ *MimeExtTable) MakeNew() *MimeExtTable {
	return DefaultMimeExtTable
}

func (mm *MimeExtTable) UnmarshalBinary(data []byte) error {
	tbl := map[string]string{}

	cur := 0
	for len(data) > 0 {
		cur++
		le := bytes.IndexByte(data, '\n')
		if le < 0 {
			return errors.New(fmt.Sprintf("not found last newline: line %d", cur))
		}

		line := data[:le]
		si := bytes.IndexByte(line, ' ')
		if si < 0 {
			return errors.New(fmt.Sprintf("bad format: line %d", cur))
		}
		if inNotPrint(line) {
			return errors.New(fmt.Sprintf("found non-printable character: line %d", cur))
		}

		ext := line[:si]
		mm_full := (line[si+1 : le])
		if bytes.ContainsRune(ext, '.') {
			return errors.New(fmt.Sprintf("bad extension: `%q`: line %d", ext, cur))
		}

		mm_typ, mm_param, err := mime.ParseMediaType(string(mm_full))
		if err != nil {
			return errors.New(fmt.Sprintf("bad mime-type: line %d", cur))
		}
		tbl[string(ext)] = mime.FormatMediaType(mm_typ, mm_param)

		data = data[le+1:]
	}

	*mm = tbl
	return nil
}
