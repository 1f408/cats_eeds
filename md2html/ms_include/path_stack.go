package ms_include

import (
	"path"
	"slices"
	"errors"
)

var ErrRecursiveInclude = errors.New("recursive include")
var ErrOverlyNestedInclude = errors.New("overly nested include")

type PathStack interface {
	Cwd() string
	Push(file string) error
	Pop()
}

const maxIncludeDepth = 100

type SlicePathStack struct {
	files []string
}

func (ps *SlicePathStack) Depth() int {
	return len(ps.files) - 1
}

func (ps *SlicePathStack) Push(file string) error {
	if ps.Depth() > maxIncludeDepth {
		return ErrOverlyNestedInclude
	}
	if ps.Contains(file) {
		return ErrRecursiveInclude
	}

	ps.files = append(ps.files, path.Clean(file))

	return nil
}

func (ps *SlicePathStack) Cwd() string {
	if len(ps.files) <= 0 {
		panic("Not initialized")
	}

	return path.Dir(ps.files[len(ps.files)-1])
}

func (ps *SlicePathStack) Pop() {
	if len(ps.files) <= 1 {
		panic("Bad pop operation")
	}

	ps.files = ps.files[:len(ps.files)-1]
}

func (ps *SlicePathStack) Contains(file string) bool {
	return slices.Contains(ps.files, file)
}

func NewSlicePathStack(start_file string) *SlicePathStack {
	return &SlicePathStack{files: []string{start_file}}
}
