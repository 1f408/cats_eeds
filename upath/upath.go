package upath

import (
	"io/fs"
	"strings"

	"github.com/l4go/unifs"
)

type UPath struct {
	p unifs.UniPath
}

var Zero = UPath{p: unifs.Zero}

func cast_upath_err(p unifs.UniPath, err error) (UPath, error) {
	if err != nil {
		return UPath{}, nil
	}
	return UPath{p: p}, err
}

func cast_upath(p unifs.UniPath) UPath {
	return UPath{p: p}
}

func New(uni_name string) (UPath, error) {
	return cast_upath_err(unifs.New(uni_name))
}

func MustNew(uni_name string) UPath {
	return UPath{p: unifs.MustNew(uni_name)}
}

func NewByOS(os_name string) (UPath, error) {
	return cast_upath_err(unifs.NewFromOSPath(os_name))
}

func MustNewByOS(os_name string) UPath {
	return UPath{p: unifs.MustNewFromOSPath(os_name)}
}

func (up UPath) IsZero() bool {
	return up.p.IsZero()
}

func (up UPath) String() string {
	return up.p.String()
}

func (up UPath) FSPath() string {
	return up.p.FSPath()
}

func FSPaths(ups []UPath) []string {
	fs_names := make([]string, len(ups))

	for i, up := range ups {
		fs_names[i] = up.FSPath()
	}

	return fs_names
}

func (up UPath) Join(names ...string) (UPath, error) {
	return cast_upath_err(up.p.Join(names...))
}

func (up UPath) MustJoin(names ...string) UPath {
	return cast_upath(up.p.MustJoin(names...))
}

func (up UPath) Open(fsys fs.FS) (fs.File, error) {
	return up.p.Open(fsys)
}

func (up UPath) Sub(fsys fs.FS) (fs.FS, error) {
	return up.p.Sub(fsys)
}

func (up UPath) Stat(fsys fs.FS) (fs.FileInfo, error) {
	return up.p.Stat(fsys)
}

func (up UPath) ReadFile(fsys fs.FS) ([]byte, error) {
	return up.p.ReadFile(fsys)
}

func (up UPath) ReadDir(fsys fs.FS) ([]fs.DirEntry, error) {
	return up.p.ReadDir(fsys)
}

func (up *UPath) UnmarshalTOML(decode func(interface{}) error) error {
	var str string
	if err := decode(&str); err != nil {
		return err
	}

	var err error = nil
	switch {
	case str == "":
		err = nil
		*up = Zero
	case strings.HasPrefix(str, "/"):
		var new_up UPath
		new_up, err = New(str)
		if err == nil {
			*up = new_up
		}
	default:
		var new_up UPath
		new_up, err = NewByOS(str)
		if err == nil {
			*up = new_up
		}
	}

	return err
}

func (up *UPath) MarshalText() ([]byte, error) {
	return []byte(up.String()), nil
}
