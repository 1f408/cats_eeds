package tmplext

import (
	"io/fs"
	"text/template"

	"github.com/1f408/cats_eeds/upath"
)

func addFunc(fmap template.FuncMap, name string, fn any) {
	if _, ok := fmap[name]; ok {
		return
	}

	fmap[name] = fn
}

func AddDefaultFunc(fmap template.FuncMap, fsys fs.FS, up upath.UPath) {
	addFunc(fmap, "url", TmplUrl)
	addFunc(fmap, "href", TmplHref)
	addFunc(fmap, "urlpath", TmplUrlPath)
	addFunc(fmap, "urlfragment", TmplUrlPath)
	addFunc(fmap, "uridata", TmplUriData)
	addFunc(fmap, "base64", TmplBase64)
	addFunc(fmap, "once", NewTmplOnce())
	addFunc(fmap, "file_type", TmplFileType)
	if fsys != nil && !up.IsZero() {
		addFunc(fmap, "svg_icon", NewTmplSvgIcon(fsys, up))
	} else {
		addFunc(fmap, "svg_icon", TmplDummySvgIcon)
	}
}
