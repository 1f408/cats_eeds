package mdview

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/l4go/rpath"
	"github.com/l4go/task"
	"github.com/l4go/unifs"

	"github.com/1f408/cats_eeds/md2html"

	"github.com/1f408/cats_eeds/view/internal/dirview"
	"github.com/1f408/cats_eeds/view/internal/etag"
	"github.com/1f408/cats_eeds/view/internal/htpath"
	"github.com/1f408/cats_eeds/view/internal/links"
)

type tmplOptions struct {
	ThemeStyle    string
	DirectoryView bool
	LocationNavi  string
}
type tmplParam struct {
	Options  *tmplOptions
	Markdown *md2html.MdConfig

	Title     string
	Top       string
	Lib       string
	Path      string
	PathLinks []links.Link
	Text      string
	TextType  string
	Toc       string
	Files     []*dirview.FileStamp
	IsOpen    bool
}

func (mdv *MdView) setCacheHeader(header Setter) {
	if mdv.CacheControl != "" {
		header.Set("Cache-Control", mdv.CacheControl)
	}
}

func set_int64bin(bin []byte, v int64) {
	binary.LittleEndian.PutUint64(bin, uint64(v))
}
func (mdv *MdView) MakeEtag(t time.Time) string {
	tm := make([]byte, 8)
	set_int64bin(tm, t.UnixMicro())

	return etag.Make(mdv.TemplateTag, tm)
}

func isModified(hd Getter, org_tag string, mod_time time.Time) bool {
	if_nmatch := hd.Get("If-None-Match")

	if if_nmatch != "" {
		return !isEtagMatch(if_nmatch, org_tag)
	}

	return true
}

func isEtagMatch(tag_str string, org_tag string) bool {
	tags, _ := etag.Split(tag_str)
	for _, tag := range tags {
		if tag == org_tag {
			return true
		}
	}

	return false
}

func (mdv *MdView) Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "405 not supported "+r.Method+" method",
			http.StatusMethodNotAllowed)
		return
	}

	mdv.writeView(r.URL.Path, r.Header, NewHttpWriter(w, r))
}

func (mdv *MdView) Dump(out, eout io.Writer, req_path string) {
	req_path = rpath.Join("/", req_path)
	h := &DummyGetter{}
	w := NewDumpWrite(out, eout)

	mdv.writeView(req_path, h, w)
}

func (mdv *MdView) writeView(req_path string, r_header Getter, w HttpWriter) {
	w_header := w.Header()

	htreq, ht_err := htpath.New(mdv.SystemFS, mdv.DocumentRoot.String(), req_path, mdv.IndexName)
	switch {
	case ht_err == nil:
	case ht_err == htpath.ErrBadRequestType:
		w.Error("400 bad request path", http.StatusBadRequest)
		return
	case os.IsNotExist(ht_err):
		w.Error("404 not found", http.StatusNotFound)
		return
	default:
		w.Error("500 file read error", http.StatusInternalServerError)
		return
	}
	if dir_mod, ok := mdv.DirViewStamp.DirModTime(htreq.Dir()); ok {
		htreq.UpdateModTime(dir_mod)
	}

	req_rpath := htreq.Req()
	is_dir := htreq.IsDir()
	has_doc := htreq.HasDoc()

	kind := htreq.Kind()
	var proc_type = ""
	var text_type = ""
	switch {
	case !has_doc && is_dir:
		proc_type = "dir"
	case kind == "text/markdown":
		proc_type = "md"
	case mdv.TextViewMode != "raw" && strings.HasPrefix(kind, "text/"):
		proc_type = "text"
		text_type = "plaintext"
	default:
		mdv.setCacheHeader(w_header)
		w.ServeFile(mdv.SystemFS, htreq.FullDoc())
		return
	}

	dir_view := true
	is_open := is_dir
	switch mdv.DirectoryViewMode {
	case "none":
		dir_view = false
		is_open = false
	case "autoindex":
		dir_view = is_dir
		is_open = true
	case "close":
		dir_view = true
		is_open = false
	case "auto":
		dir_view = true
		is_open = is_dir
	case "open":
		dir_view = true
		is_open = true
	default:
		w.Error("503 bad directory view mode",
			http.StatusServiceUnavailable)
		return
	}

	mod_time := htreq.ModTime()
	if mod_time.Before(mdv.ConfigModTime) {
		mod_time = mdv.ConfigModTime
	}
	last_mod := htreq.LastMod()

	tag := mdv.MakeEtag(mod_time)
	if !isModified(r_header, tag, mod_time) {
		w_header.Set("Last-Modified", last_mod)
		w_header.Set("Etag", tag)
		w.WriteHeader(http.StatusNotModified)
		return
	}

	var raw_bin []byte
	if has_doc {
		var rd_err error
		raw_bin, rd_err = unifs.ReadFile(mdv.SystemFS, htreq.FullDoc())
		if rd_err != nil {
			w.Error("500 document file read error",
				http.StatusInternalServerError)
			return
		}
	}

	var doc_bin []byte
	var title_bin []byte
	var toc_bin []byte
	req_abs_path := rpath.Join(mdv.UrlTopPath, req_rpath)

	switch proc_type {
	default:
		w.Error("500 media handling error", http.StatusInternalServerError)
		return
	case "dir":
		doc_bin = []byte{}
		toc_bin = []byte{}
		title_bin = []byte("View: " + req_abs_path)
	case "text":
		doc_bin = raw_bin
		toc_bin = []byte{}
		title_bin = []byte("View: " + req_abs_path)
	case "md":
		var cerr error
		doc_bin, toc_bin, title_bin, cerr = mdv.Md2Html.Convert(raw_bin)
		if cerr != nil {
			w.Error("500 conversion failed", http.StatusInternalServerError)
			return
		}
		if len(title_bin) == 0 {
			title_bin = []byte("View: " + req_abs_path)
		}
	}

	var f_list []*dirview.FileStamp = nil
	if dir_view {
		f_list = mdv.DirViewStamp.Get(htreq.Dir(), !is_dir)
	}

	tmpl_param := tmplParam{
		Options: &tmplOptions{
			ThemeStyle:    mdv.ThemeStyle,
			DirectoryView: dir_view,
			LocationNavi:  mdv.LocationNavi,
		},
		Markdown:  mdv.MarkdownConfig,
		Top:       mdv.UrlTopPath,
		Lib:       mdv.UrlLibPath,
		Path:      req_abs_path,
		PathLinks: links.NewLinks(rpath.Join("/", req_rpath)),
		Text:      string(doc_bin),
		TextType:  text_type,
		Title:     string(title_bin),
		Toc:       string(toc_bin),
		Files:     f_list,
		IsOpen:    is_open,
	}

	var buf bytes.Buffer
	err := mdv.execTemplate(&buf, tmpl_param)
	if err != nil {
		w.Error("503 template execute error:"+err.Error(),
			http.StatusServiceUnavailable)
		return
	}

	w_header.Set("Content-Type", "text/html; charset=utf-8")
	w_header.Set("Last-Modified", last_mod)
	w_header.Set("Etag", tag)
	mdv.setCacheHeader(w_header)
	buf.WriteTo(w)
}

func (mdv *MdView) execTemplate(w io.Writer, param interface{}) error {
	tmpl, err := mdv.OriginTmpl.Clone()
	if err != nil {
		return err
	}

	tmpl_funcs := template.FuncMap{
		"once":      NewTmplOnce(),
		"svg_icon":  mdv.TmplSvgIcon,
		"file_type": TmplFileType,
	}
	tmpl = tmpl.Funcs(tmpl_funcs)

	return tmpl.ExecuteTemplate(w, mdv.MainTmplName, param)
}

func (mdv *MdView) SumTemplate() ([]byte, error) {
	tmpl_param := tmplParam{
		Options: &tmplOptions{
			ThemeStyle:    mdv.ThemeStyle,
			DirectoryView: (mdv.DirectoryViewMode != "none"),
			LocationNavi:  mdv.LocationNavi,
		},
		Markdown:  mdv.MarkdownConfig,
		Top:       mdv.UrlTopPath,
		Lib:       mdv.UrlLibPath,
		Path:      mdv.UrlTopPath,
		PathLinks: links.NewLinks("/"),
		Text:      "",
		TextType:  "md",
		Title:     "",
		Toc:       "",
		Files:     nil,
		IsOpen:    false,
	}

	h_ctx := sha256.New()
	err := mdv.execTemplate(h_ctx, tmpl_param)
	if err != nil {
		return nil, err
	}

	return h_ctx.Sum(nil), nil
}

var ErrUnsupportedSocketType = errors.New("unsupported socket type.")

func listen(cc task.Canceller, stype string, spath string) (net.Listener, error) {
	lcnf := &net.ListenConfig{}

	switch stype {
	default:
		return nil, ErrUnsupportedSocketType
	case "unix":
	case "tcp":
	}

	return lcnf.Listen(cc.AsContext(), stype, spath)
}

func (mdv *MdView) ListenAndServe(cc task.Canceller) error {
	lstn, lerr := listen(cc, mdv.SocketType, mdv.SocketPath)
	switch lerr {
	case nil:
	case context.Canceled:
	default:
		return new_err("socket listen error: %v.", lerr)
	}

	if mdv.SocketType == "unix" {
		defer os.Remove(mdv.SocketPath)
		os.Chmod(mdv.SocketPath, 0777)
	}

	return mdv.Serve(cc, lstn)
}

func (mdv *MdView) Serve(cc task.Canceller, lstn net.Listener) error {
	if mdv.SocketType == "" || mdv.SocketPath == "" {
		addr := lstn.Addr()
		mdv.SocketType = addr.Network()
		mdv.SocketPath = addr.String()
	}

	srv := &http.Server{Addr: mdv.SocketPath}
	go func() {
		select {
		case <-cc.RecvCancel():
		}
		srv.Close()
	}()

	http.HandleFunc("/", mdv.Handler)

	serr := srv.Serve(lstn)
	switch serr {
	default:
		return new_err("HTTP server error: %v.", serr)
	case nil:
	case http.ErrServerClosed:
	}

	return nil
}
