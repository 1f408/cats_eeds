package upath

import (
	"encoding"
	"io"
	"io/fs"
	"time"

	"github.com/naoina/toml"
)

type NewMaker[T any] interface {
	MakeNew() T
}

type Import[T NewMaker[T]] struct {
	UPath   UPath
	ModTime time.Time `toml:"-"`
	Value   T         `toml:"-"`
}

func (im *Import[T]) UnmarshalTOML(decode func(interface{}) error) error {
	return decode(&im.UPath)
}

func decodeFile(vi any, f fs.File) error {
	if v, ok := vi.(toml.UnmarshalerRec); ok {
		return toml.NewDecoder(f).Decode(v)
	}

	if v, ok := vi.(encoding.BinaryUnmarshaler); ok {
		text, err := io.ReadAll(f)
		if err != nil {
			return err
		}
		return v.UnmarshalBinary(text)
	}

	return toml.NewDecoder(f).Decode(vi)
}

func (im *Import[T]) decodeFS(fsys fs.FS) error {
	f, err := im.UPath.Open(fsys)
	if err != nil {
		return &fs.PathError{Op: "import", Path: im.UPath.String(), Err: err}
	}

	if err := decodeFile(im.Value, f); err != nil {
		return &fs.PathError{Op: "import", Path: im.UPath.String(), Err: err}
	}

	if fi, err := f.Stat(); err != nil {
		im.ModTime = fi.ModTime()
	}

	return nil
}

func (im *Import[T]) RebuildByType(fsys fs.FS) error {
	im.Value = im.Value.MakeNew()

	if im.UPath.IsZero() {
		return nil
	}

	if err := im.decodeFS(fsys); err != nil {
		return err
	}

	return nil
}
