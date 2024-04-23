package tmplview

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
type tmplHtmlParam struct {
	Options  *tmplOptions
	Markdown *md2html.MdConfig

	Title     string
	Top       string
	Lib       string
	Path      string
	PathLinks []links.Link
	Files  []*dirview.FileStamp
	IsOpen bool

	UserName string
}

type tmplMdParam struct {
	Options  *tmplOptions
	Markdown *md2html.MdConfig

	Title     string
	Top       string
	Lib       string
	Path      string
	PathLinks []links.Link
	Files  []*dirview.FileStamp
	IsOpen bool

	Text      string
	TextType  string
	Toc       string
}

func (tmpv *TmplView) setCacheHeader(header Setter) {
	if tmpv.CacheControl != "" {
		header.Set("Cache-Control", tmpv.CacheControl)
	}
}

func set_int64bin(bin []byte, v int64) {
	binary.LittleEndian.PutUint64(bin, uint64(v))
}

func (tmpv *TmplView) MakeEtag(t time.Time, user string) string {
	tm := make([]byte, 8)
	set_int64bin(tm, t.UnixMicro())

	return etag.Make(tmpv.TemplateTag, tm, etag.Crypt(tm, []byte(user)))
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

func (tmpv *TmplView) Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "405 not supported "+r.Method+" method",
			http.StatusMethodNotAllowed)
		return
	}

	tmpv.writeView(r.URL.Path, r.Header, NewHttpWriter(w, r))
}

func (tmpv *TmplView) Dump(out, eout io.Writer, req_path string) {
	req_path = rpath.Join("/", req_path)
	h := &DummyGetter{}
	w := NewDumpWrite(out, eout)

	tmpv.writeView(req_path, h, w)
}

func (tmpv *TmplView) writeView(req_path string, r_header Getter, w HttpWriter) {
	w_header := w.Header()
	htreq, ht_err := htpath.New(tmpv.SystemFS, tmpv.DocumentRoot.String(),
		req_path, tmpv.IndexName)
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
	if dir_mod, ok := tmpv.DirViewStamp.DirModTime(htreq.Dir()); ok {
		htreq.UpdateModTime(dir_mod)
	}

	user := r_header.Get(tmpv.AuthnUserHeader)

	req_rpath := htreq.Req()
	is_dir := htreq.IsDir()
	has_doc := htreq.HasDoc()

	dir_view := true
	is_open := is_dir
	switch tmpv.DirectoryViewMode {
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
	}

	kind := htreq.Kind()
	mime := htreq.Mime()

	var proc_type string = ""
	var text_type string = ""
	switch {
	case !has_doc && is_dir:
		proc_type = "dir"
	case kind == "text/html":
		proc_type = "html"
	case kind == "text/markdown" && tmpv.MdTmplName != "":
		proc_type = "md"
	case tmpv.TextViewMode != "raw" && strings.HasPrefix(kind, "text/"):
		proc_type = "text"
		text_type = "plaintext"
	default:
		tmpv.setCacheHeader(w_header)
		w.ServeFile(tmpv.SystemFS, htreq.FullDoc())
		return
	}

	switch proc_type {
	default:
		w.Error("415 unsupported media type",
			http.StatusUnsupportedMediaType)
		return
	case "html":
	case "text":
	case "md":
	case "dir":
	}

	mod_time := htreq.ModTime()
	if mod_time.Before(tmpv.ConfigModTime) {
		mod_time = tmpv.ConfigModTime
	}
	last_mod := htreq.LastMod()

	tag := tmpv.MakeEtag(mod_time, user)
	if !isModified(r_header, tag, mod_time) {
		w_header.Set("Last-Modified", last_mod)
		w_header.Set("Etag", tag)
		w.WriteHeader(http.StatusNotModified)
		return
	}

	var raw_bin []byte
	if has_doc {
		var rd_err error
		raw_bin, rd_err = unifs.ReadFile(tmpv.SystemFS, htreq.FullDoc())
		if rd_err != nil && !os.IsNotExist(rd_err) {
			w.Error("500 document file read error",
				http.StatusInternalServerError)
			return
		}
	}

	tmpl, err := tmpv.OriginTmpl.Clone()
	if err != nil {
		w.Error("503 service unavailable: "+err.Error(),
			http.StatusServiceUnavailable)
		return
	}

	req_abs_path := rpath.Join(tmpv.UrlTopPath, req_rpath)
	var f_list []*dirview.FileStamp = nil
	if dir_view {
		f_list = tmpv.DirViewStamp.Get(htreq.Dir(), !is_dir)
	}

	tmpl_funcs := template.FuncMap{
		"once": NewTmplOnce(),
		"in_group": func(grp string) bool {
			return tmpv.UserMap.InGroup(user, grp)
		},
		"in_user": func() bool {
			return tmpv.UserMap.InUser(user)
		},
	}
	if proc_type != "text" {
		tmpl_funcs["svg_icon"] = tmpv.TmplSvgIcon
		tmpl_funcs["file_type"] = TmplFileType
	}

	tmpl = tmpl.Funcs(tmpl_funcs)

	tmpl, err = tmpl.Parse(string(raw_bin))
	if err != nil {
		w.Error("503 service unavailable: "+err.Error(),
			http.StatusServiceUnavailable)
		return
	}

	tmpl_param := tmplHtmlParam{
		Options: &tmplOptions{
			ThemeStyle:    tmpv.ThemeStyle,
			DirectoryView: dir_view,
			LocationNavi:  tmpv.LocationNavi,
		},
		Markdown: tmpv.MarkdownConfig,

		Top:       tmpv.UrlTopPath,
		Lib:       tmpv.UrlLibPath,
		Path:      req_abs_path,
		PathLinks: links.NewLinks(rpath.Join("/", req_rpath)),
		Files:     f_list,
		IsOpen:    is_open,

		UserName: user,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, tmpl_param); err != nil {
		w.Error("500 html template execute error:"+err.Error(),
			http.StatusInternalServerError)
		return
	}

	var doc_bin []byte
	var toc_bin []byte
	var title_bin []byte

	switch proc_type {
	default:
		w.Error("500 media handling error", http.StatusInternalServerError)
		return
	case "html":
		w_header.Set("Content-Type", mime)
		w_header.Set("Last-Modified", last_mod)
		w_header.Set("Etag", tag)
		tmpv.setCacheHeader(w_header)
		buf.WriteTo(w)
		return

	case "dir":
		doc_bin = []byte{}
		toc_bin = []byte{}
		title_bin = []byte("View: " + req_abs_path)
	case "text":
		doc_bin = buf.Bytes()
		toc_bin = []byte{}
		title_bin = []byte("View: " + req_abs_path)
	case "md":
		var cerr error
		doc_bin, toc_bin, title_bin, cerr = tmpv.Md2Html.Convert(buf.Bytes())
		if cerr != nil {
			w.Error("500 conversion failed", http.StatusInternalServerError)
			return
		}
		if len(title_bin) == 0 {
			title_bin = []byte("View: " + req_abs_path)
		}
	}

	mdtmpl_param := tmplMdParam{
		Options: &tmplOptions{
			ThemeStyle:    tmpv.ThemeStyle,
			DirectoryView: dir_view,
			LocationNavi:  tmpv.LocationNavi,
		},
		Markdown: tmpv.MarkdownConfig,

		Title:     string(title_bin),
		Top:       tmpv.UrlTopPath,
		Lib:       tmpv.UrlLibPath,
		Path:      req_abs_path,
		PathLinks: links.NewLinks(rpath.Join("/", req_rpath)),
		Files:     f_list,
		IsOpen:    is_open,

		Text:      string(doc_bin),
		TextType:  text_type,
		Toc:       string(toc_bin),
	}
	var mdbuf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&mdbuf, tmpv.MdTmplName, mdtmpl_param); err != nil {
		w.Error("500 markdown template execute error:"+err.Error(),
			http.StatusInternalServerError)
		return
	}

	w_header.Set("Content-Type", "text/html; charset=UTF-8")
	w_header.Set("Last-Modified", last_mod)
	w_header.Set("Etag", tag)
	tmpv.setCacheHeader(w_header)
	mdbuf.WriteTo(w)
}

func (tmpv *TmplView) SumTemplate() ([]byte, error) {
	tmpl, err := tmpv.OriginTmpl.Clone()
	if err != nil {
		return nil, err
	}

	tmpl_funcs := template.FuncMap{
		"once":      NewTmplOnce(),
		"svg_icon":  tmpv.TmplSvgIcon,
		"file_type": TmplFileType,
		"in_group": func(grp string) bool {
			return true
		},
		"in_user": func() bool {
			return true
		},
	}
	tmpl = tmpl.Funcs(tmpl_funcs)

	param := tmplMdParam{
		Options: &tmplOptions{
			ThemeStyle:    tmpv.ThemeStyle,
			DirectoryView: (tmpv.DirectoryViewMode != "none"),
			LocationNavi:  tmpv.LocationNavi,
		},
		Markdown:  tmpv.MarkdownConfig,
		Top:       tmpv.UrlTopPath,
		Lib:       tmpv.UrlLibPath,
		Path:      tmpv.UrlTopPath,
		PathLinks: links.NewLinks("/"),
		Text:      "",
		TextType:  "md",
		Title:     "",
		Toc:       "",
		Files:     nil,
		IsOpen:    false,
	}

	h_ctx := sha256.New()
	if e := tmpv.WriteTestCatUi(h_ctx); e != nil {
		return nil, e
	}
	if e := tmpl.ExecuteTemplate(h_ctx, tmpv.MdTmplName, param); e != nil {
		return nil, e
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

func (tmpv *TmplView) ListenAndServe(cc task.Canceller) error {
	lstn, lerr := listen(cc, tmpv.SocketType, tmpv.SocketPath)
	switch lerr {
	case nil:
	case context.Canceled:
	default:
		return new_err("socket listen error: %v.", lerr)
	}

	if tmpv.SocketType == "unix" {
		defer os.Remove(tmpv.SocketPath)
		os.Chmod(tmpv.SocketPath, 0777)
	}

	return tmpv.Serve(cc, lstn)
}

func (tmpv *TmplView) Serve(cc task.Canceller, lstn net.Listener) error {
	srv := &http.Server{Addr: tmpv.SocketPath}
	go func() {
		select {
		case <-cc.RecvCancel():
		}
		srv.Close()
	}()

	if tmpv.SocketType == "" || tmpv.SocketPath == "" {
		addr := lstn.Addr()
		tmpv.SocketType = addr.Network()
		tmpv.SocketPath = addr.String()
	}

	http.HandleFunc("/", tmpv.Handler)

	serr := srv.Serve(lstn)
	switch serr {
	default:
		return new_err("HTTP server error: %v.", serr)
	case nil:
	case http.ErrServerClosed:
	}

	return nil
}
