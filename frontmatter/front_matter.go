package frontmatter

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

var ErrNotFound = errors.New("not found")

func extract_word(ln []byte) string {
	ln = bytes.TrimSuffix(ln, []byte{'\n'})
	ln = bytes.TrimSuffix(ln, []byte{'\r'})
	return string(ln)
}

type FrontMatter struct {
	wd_tbl map[string]struct{}
}

func New(words []string) *FrontMatter {
	wd_tbl := map[string]struct{}{}
	for _, w := range words {
		wd_tbl[string(w)] = struct{}{}
	}

	return &FrontMatter{wd_tbl: wd_tbl}
}

func (fm *FrontMatter) FindAndSplit(r io.Reader) (string, []byte, []byte, error) {
	p := fm.newParser(r)
	wd, err := p.find_begin()
	if err != nil {
		return "", nil, nil, err
	}

	if err := p.seek_end(wd); err != nil {
		return "", nil, nil, err
	}

	if err := p.seek_eof(); err != nil {
		return "", nil, nil, err
	}

	return wd, p.head(), p.body(), nil
}

type parser struct {
	fm       *FrontMatter
	r        *bufio.Reader
	w        *bytes.Buffer
	hd_begin int
	hd_end   int
	bd_begin int
}

func (fm *FrontMatter) newParser(r io.Reader) *parser {
	return &parser{
		fm:       fm,
		r:        bufio.NewReader(r),
		w:        bytes.NewBuffer(nil),
		hd_begin: 0,
		hd_end:   0,
		bd_begin: 0,
	}
}

func (p *parser) head() []byte {
	return p.w.Bytes()[p.hd_begin:p.hd_end]
}

func (p *parser) body() []byte {
	return p.w.Bytes()[p.bd_begin:]
}

func (p *parser) find_begin() (string, error) {
	for {
		ln, err := p.seek_line()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		p.hd_begin += len(ln)
		wd := extract_word(ln)

		_, ok := p.fm.wd_tbl[wd]
		if ok {
			p.hd_end = p.hd_begin
			return wd, nil
		}

		break
	}

	return "", ErrNotFound
}

func (p *parser) seek_end(end_wd string) error {
	is_eof := false
	for !is_eof {
		ln, err := p.seek_line()
		is_eof = err == io.EOF
		if err != nil && !is_eof {
			return err
		}

		wd := extract_word(ln)
		if wd == end_wd {
			p.bd_begin = p.hd_end + len(ln)
			return nil
		}

		p.hd_end += len(ln)
	}

	return ErrNotFound
}

func (p *parser) seek_eof() error {
	_, err := p.w.ReadFrom(p.r)
	return err
}

func (p *parser) seek_line() ([]byte, error) {
	ln, rerr := p.r.ReadBytes('\n')
	if rerr != nil && rerr != io.EOF {
		return nil, rerr
	}

	if _, werr := p.w.Write(ln); werr != nil {
		return nil, werr
	}

	return ln, rerr
}
