package htpath

import (
	"errors"
	"io/fs"
	"net/http"
	"path"
	"time"

	"github.com/l4go/rpath"
	"github.com/l4go/unifs"
)

type HttpPath struct {
	fsys   fs.FS
	root   string
	req    string
	is_dir bool
	index  string

	dir  string
	file string
	ext  string
	kind string
	mime string

	mod_time time.Time
}

var ErrBadRequestType = errors.New("bad request type")
var ErrBadIndexBase = errors.New("bad index base")

func New(fsys fs.FS, root string, req string, index string) (*HttpPath, error) {
	root = rpath.SetDir(root)
	req = rpath.Clean(req)
	index = rpath.Clean(index)
	if rpath.IsDir(index) {
		return nil, ErrBadIndexBase
	}

	return new_by_rpath(fsys, root, req, index)
}

func (hp *HttpPath) NewSibling(relpath string) (*HttpPath, error) {
	req := rpath.Join(hp.dir, relpath)

	return new_by_rpath(hp.fsys, hp.root, req, hp.index)
}

func new_by_rpath(fsys fs.FS, root string, req string, index string) (*HttpPath, error) {
	var dir string
	var file string

	is_dir := rpath.IsDir(req)

	req_fi, req_err := unifs.Stat(fsys, path.Join(root, req))
	if req_err != nil {
		return nil, req_err
	}
	if req_fi.IsDir() {
		if !is_dir {
			req = rpath.SetDir(req)
			is_dir = true
		}
	} else if is_dir {
		return nil, ErrBadRequestType
	}
	mod_time := req_fi.ModTime()

	if is_dir {
		dir = req
		file = ""
		if index != "" {
			fi, err := unifs.Stat(fsys, path.Join(root, dir, index))
			if err == nil && !fi.IsDir() {
				file = index

				idx_mod_time := fi.ModTime()
				if idx_mod_time.After(mod_time) {
					mod_time = idx_mod_time
				}
			}
		}
	} else {
		dir, file = rpath.Split(req)
	}

	ext := rpath.Ext(file)
	kind := ""
	mime := ""
	if ext != "" {
		kind, mime = GetFileKindByExt(ext)
	}

	return &HttpPath{
		root:   root,
		req:    req,
		is_dir: is_dir,
		index:  index,

		dir:  dir,
		file: file,
		ext:  ext,
		kind: kind,
		mime: mime,

		mod_time: mod_time,
	}, nil
}

func (hp *HttpPath) Req() string {
	return hp.req
}

func (hp *HttpPath) FullReq() string {
	return rpath.Join(hp.root, hp.req)
}

func (hp *HttpPath) IsDir() bool {
	return hp.is_dir
}

func (hp *HttpPath) HasDoc() bool {
	return hp.file != ""
}

func (hp *HttpPath) UpdateModTime(mod time.Time) {
	if !mod.IsZero() && hp.mod_time.Before(mod) {
		hp.mod_time = mod
	}
}

func (hp *HttpPath) ModTime() time.Time {
	return hp.mod_time
}

func (hp *HttpPath) LastMod() string {
	return hp.mod_time.Format(http.TimeFormat)
}

func (hp *HttpPath) Root() string {
	return hp.root
}

func (hp *HttpPath) Dir() string {
	return hp.dir
}

func (hp *HttpPath) FullDir() string {
	return rpath.Join(hp.root, hp.dir)
}

func (hp *HttpPath) Name() string {
	return hp.file
}

func (hp *HttpPath) Doc() string {
	if hp.file == "" {
		return ""
	}
	return rpath.Join(hp.dir, hp.file)
}

func (hp *HttpPath) FullDoc() string {
	if hp.file == "" {
		return ""
	}
	return rpath.Join(hp.root, hp.dir, hp.file)
}

func (hp *HttpPath) Ext() string {
	return hp.ext
}
func (hp *HttpPath) Kind() string {
	return hp.kind
}
func (hp *HttpPath) Mime() string {
	return hp.mime
}
