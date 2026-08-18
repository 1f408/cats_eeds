package mdview

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
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

	"github.com/1f408/cats_eeds/frontmatter"
	"github.com/1f408/cats_eeds/internal/perenc"
	"github.com/1f408/cats_eeds/md2html"
	"github.com/1f408/cats_eeds/view/internal/dirview"
	"github.com/1f408/cats_eeds/view/internal/etag"
	"github.com/1f408/cats_eeds/view/internal/htpath"
	"github.com/1f408/cats_eeds/view/internal/links"
	"github.com/1f408/cats_eeds/view/internal/tmplext"
)

type tmplParam struct {
	Options  *tmplOptions
	Markdown *md2html.MdConfig
	SmCard   *md2html.SmCardParam

	Title     string
	Top       string
	Lib       string
	Path      string
	PathLinks []links.Link
	LinkMenu  []md2html.Link
	Text      string
	TextType  string
	Toc       string
	Files     []*dirview.FileStamp
	IsOpen    bool

	CustomParam md2html.CustomParam
}

type tmplOptions struct {
	ThemeStyle   string
	PageStyle    string
	PrintSizeCss string
	PrintZoom    float32

	LocationNavi  string
	TocNavi       string
	DirectoryView bool
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

	req_path := rpath.Clean("/" + r.URL.Path)
	mdv.writeView(req_path, r.Header, NewHttpWriter(w, r))
}

func (mdv *MdView) Dump(out, eout io.Writer, req_path string) {
	h := &DummyGetter{}
	w := NewDumpWrite(out, eout)

	req_path = rpath.Clean("/" + req_path)
	mdv.writeView(req_path, h, w)
}

func (mdv *MdView) writeView(req_path string, r_header Getter, w HttpWriter) {
	w_header := w.Header()

	htreq, ht_err := htpath.New(mdv.SystemFS, mdv.DocumentRoot.String(), req_path, mdv.IndexName)
	switch {
	case ht_err == nil:
	case ht_err == htpath.ErrInvalidDirRequest:
		if !mdv.DirectoryRedirection {
			w.Error("404 invalid directory URL", http.StatusNotFound)
			return
		}

		p := rpath.SetDir(rpath.Join(mdv.UrlTopPath, req_path))
		w_header.Set("Location", perenc.EncodeUrlPath(p))
		w.Error("307 use the directory URL", http.StatusTemporaryRedirect)
		return
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
	var fm_param *md2html.FrontMatterParam = &md2html.FrontMatterParam{}
	if has_doc {
		var rd_err error
		raw_bin, rd_err = unifs.ReadFile(mdv.SystemFS, htreq.FullDoc())
		if rd_err != nil {
			w.Error("500 document file read error",
				http.StatusInternalServerError)
			return
		}

		if (proc_type == "md" ||
			proc_type == "text" && mdv.CustomPageConfig.FrontMatter.UsedForText) &&
			mdv.CustomPageConfig.FrontMatter.IsEnabled() {

			body, fmp, fm_err := mdv.CustomPageConfig.FrontMatter.TrimAndParse(raw_bin)
			switch fm_err {
			case nil:
				if fmp == nil {
					w.Error("500 Frontmatter parse error", http.StatusInternalServerError)
					return
				}
				raw_bin = body
				fm_param = fmp
			case frontmatter.ErrNotFound:
			default:
				w.Error("500 Frontmatter error: "+fm_err.Error(),
					http.StatusInternalServerError)
				return
			}
		}
	}

	theme_style := mdv.ThemeStyle
	if fm_param.ThemeStyle != "" {
		theme_style = fm_param.ThemeStyle
	}

	page_style := mdv.PageStyle
	if fm_param.PageStyle != "" {
		page_style = fm_param.PageStyle
	}
	style_tmpl := ""
	if page_style != "" {
		style_tmpl = "style_" + page_style + ".tmpl"
	}

	page_size_css := mdv.PrintSizeCss
	if fm_param.PaperType != "" {
		css, ok := mdv.PrintPaperMapping.GetCss(fm_param.PaperType)
		if ok {
			page_size_css = css
		}
	}
	print_zoom := mdv.PrintZoom
	if fm_param.PrintZoom > 0 {
		print_zoom = fm_param.PrintZoom
	}

	loc_navi := mdv.LocationNavi
	if fm_param.LocationNavi != "" {
		loc_navi = fm_param.LocationNavi
	}

	toc_navi := mdv.TocNavi
	if fm_param.TocNavi != "" {
		toc_navi = fm_param.TocNavi
	}

	dir_view_mode := mdv.DirectoryViewMode
	if fm_param.DirectoryViewMode != "" {
		dir_view_mode = fm_param.DirectoryViewMode
	}

	dir_view := true
	is_open := is_dir
	switch dir_view_mode {
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

	tmpl, err := mdv.OriginTmpl.Clone()
	if err != nil {
		w.Error("503 service unavailable: "+err.Error(),
			http.StatusServiceUnavailable)
		return
	}

	var doc_bin []byte
	var title_bin []byte
	var toc_bin []byte
	req_abs_path := rpath.Join(mdv.UrlTopPath, req_rpath)

	with_title_param := false
	title_bin = []byte("View: " + req_abs_path)
	if fm_param.Title != "" {
		with_title_param = true
		title_bin = []byte(fm_param.Title)
	}

	var sm_card *md2html.SmCardParam = nil
	if mdv.CustomPageConfig.SmCard.Enabled {
		sm_card = &md2html.SmCardParam{}
		*sm_card = fm_param.SmCard
		sm_card.Fix(&mdv.CustomPageConfig.SmCard, req_abs_path)

		if sm_card.Title == "" {
			sm_card.Title = string(title_bin)
		}
	}

	switch proc_type {
	default:
		w.Error("500 media handling error", http.StatusInternalServerError)
		return
	case "dir":
		doc_bin = []byte{}
		toc_bin = []byte{}
	case "text":
		doc_bin = raw_bin
		toc_bin = []byte{}
	case "md":
		var cerr error
		var md_title_bin []byte

		m2h := md2html.NewMd2Html(&md2html.Md2HtmlConfig{
			MdConfig:    mdv.MarkdownConfig,
			SystemIds:   mdv.SystemHtmlIds,
			SystemFS:    mdv.SystemFS,
			FrontMatter: mdv.CustomPageConfig.FrontMatter,
			StartMdFile: htreq.FullDoc(),
		})

		if fm_param.MarkdownConfig != "" {
			name := fm_param.MarkdownConfig
			if name[0] != '/' {
				name = rpath.Join(rpath.Dir(htreq.FullDoc()), name)
			}
			if strings.HasPrefix(name, mdv.DocumentRoot.String()) {
				if md_cfg, err := md2html.NewMdConfig(mdv.SystemFS, name); err == nil {
					m2h = m2h.NewLocalSpec(md_cfg)
				}
			}
		}

		doc_bin, toc_bin, md_title_bin, cerr = m2h.Convert(raw_bin)
		if cerr != nil {
			w.Error("500 conversion failed: "+cerr.Error(), http.StatusInternalServerError)
			return
		}

		if !with_title_param {
			title_bin = md_title_bin
			if sm_card != nil && sm_card.Title == "" {
				sm_card.Title = string(md_title_bin)
			}
		}
	}

	var f_list []*dirview.FileStamp = nil
	if dir_view {
		f_list = mdv.DirViewStamp.Get(htreq.Dir(), !is_dir)
	}

	link_menu := []md2html.Link{}
	if mdv.CustomPageConfig.LinkMenu.Default != nil {
		link_menu = mdv.CustomPageConfig.LinkMenu.Default
	}
	if fm_param.LinkMenu != nil {
		link_menu = fm_param.LinkMenu
	}

	custom_param := md2html.CustomParam{}
	if mdv.CustomPageConfig.CustomParam.Default != nil {
		custom_param = mdv.CustomPageConfig.CustomParam.Default
	}
	if fm_param.CustomParam != nil {
		custom_param = fm_param.CustomParam
	}

	tmpl_param := tmplParam{
		Options: &tmplOptions{
			ThemeStyle:    theme_style,
			PageStyle:     page_style,
			PrintSizeCss:  page_size_css,
			PrintZoom:     print_zoom,
			LocationNavi:  loc_navi,
			TocNavi:       toc_navi,
			DirectoryView: (dir_view_mode != "none"),
		},
		Markdown: mdv.MarkdownConfig,
		SmCard:   sm_card,

		Top:       mdv.UrlTopPath,
		Lib:       mdv.UrlLibPath,
		Path:      req_abs_path,
		PathLinks: links.NewLinks(rpath.Join("/", req_rpath)),
		LinkMenu:  link_menu,
		Text:      string(doc_bin),
		TextType:  text_type,
		Title:     string(title_bin),
		Toc:       string(toc_bin),
		Files:     f_list,
		IsOpen:    is_open,

		CustomParam: custom_param,
	}

	tmpl = tmplLookups(tmpl, style_tmpl, mdv.MainTmplName)
	if tmpl == nil {
		w.Error("503 not found template", http.StatusServiceUnavailable)
		return
	}

	tmpl_funcs := template.FuncMap{}
	tmplext.AddDefaultFunc(tmpl_funcs, mdv.SystemFS, mdv.SvgIconPath)
	tmpl = tmpl.Funcs(tmpl_funcs)

	var buf bytes.Buffer
	if e := tmpl.Execute(&buf, tmpl_param); e != nil {
		w.Error("503 template execute error:"+e.Error(),
			http.StatusServiceUnavailable)
		return
	}

	w_header.Set("Content-Type", "text/html; charset=utf-8")
	w_header.Set("Last-Modified", last_mod)
	w_header.Set("Etag", tag)
	mdv.setCacheHeader(w_header)
	buf.WriteTo(w)
}

func tmplLookups(tmpl *template.Template, names ...string) *template.Template {
	var tt *template.Template = nil
	for _, n := range names {
		if n != "" {
			tt = tmpl.Lookup(n)
			if tt != nil {
				break
			}
		}
	}

	return tt
}

type TmplSum struct {
	Sha256    []byte
	SystemIds []string
}

func (mdv *MdView) SumTemplate() (*TmplSum, error) {
	tmpl, cerr := mdv.OriginTmpl.Clone()
	if cerr != nil {
		return nil, cerr
	}
	tmpl = tmpl.Lookup(mdv.MainTmplName)
	if tmpl == nil {
		return nil, fmt.Errorf("template: no template %q", mdv.MainTmplName)
	}

	tmpl_param := tmplParam{
		Options: &tmplOptions{
			ThemeStyle:    mdv.ThemeStyle,
			PageStyle:     mdv.PageStyle,
			PrintSizeCss:  mdv.PrintSizeCss,
			PrintZoom:     mdv.PrintZoom,
			LocationNavi:  mdv.LocationNavi,
			TocNavi:       mdv.TocNavi,
			DirectoryView: (mdv.DirectoryViewMode != "none"),
		},
		Markdown:  mdv.MarkdownConfig,
		Top:       mdv.UrlTopPath,
		Lib:       mdv.UrlLibPath,
		Path:      mdv.UrlTopPath,
		PathLinks: links.NewLinks("/"),
		LinkMenu:  nil,
		Text:      "TEST text\n",
		TextType:  "md",
		Title:     "TEST title",
		Toc:       "<ul></ul>",
		Files:     nil,
		IsOpen:    false,

		CustomParam: md2html.CustomParam{},
	}

	var b bytes.Buffer
	h_ctx := sha256.New()
	all_w := io.MultiWriter(&b, h_ctx)

	if e := tmpl.Execute(all_w, tmpl_param); e != nil {
		return nil, e
	}
	sha256_sum := h_ctx.Sum(nil)

	ids, fe := md2html.FindHtmlIds(bytes.NewReader(b.Bytes()))
	if fe != nil {
		return nil, fe
	}

	return &TmplSum{Sha256: sha256_sum, SystemIds: ids}, nil
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

	srv := &http.Server{Addr: mdv.SocketPath, Handler: http.HandlerFunc(mdv.Handler)}
	go func() {
		select {
		case <-cc.RecvCancel():
		}
		srv.Close()
	}()

	serr := srv.Serve(lstn)
	switch serr {
	default:
		return new_err("HTTP server error: %v.", serr)
	case nil:
	case http.ErrServerClosed:
	}

	return nil
}
