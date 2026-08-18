package tmplview

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

type tmplOptions struct {
	ThemeStyle    string
	PageStyle     string
	PrintSizeCss  string
	PrintZoom     float32
	LocationNavi  string
	TocNavi       string
	DirectoryView bool
}

type tmplHtmlParam struct {
	Options  *tmplOptions
	Markdown *md2html.MdConfig
	SmCard   *md2html.SmCardParam

	Title     string
	Top       string
	Lib       string
	Path      string
	PathLinks []links.Link
	LinkMenu  []md2html.Link
	Files     []*dirview.FileStamp
	IsOpen    bool

	UserName string

	CustomParam md2html.CustomParam
}

type tmplMdParam struct {
	Options  *tmplOptions
	Markdown *md2html.MdConfig
	SmCard   *md2html.SmCardParam

	Title     string
	Top       string
	Lib       string
	Path      string
	PathLinks []links.Link
	LinkMenu  []md2html.Link
	Files     []*dirview.FileStamp
	IsOpen    bool

	Text     string
	TextType string
	Toc      string

	CustomParam md2html.CustomParam
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

	req_path := rpath.Clean("/" + r.URL.Path)
	tmpv.writeView(req_path, r.Header, NewHttpWriter(w, r))
}

func (tmpv *TmplView) Dump(out, eout io.Writer, req_path string) {
	h := &DummyGetter{}
	w := NewDumpWrite(out, eout)

	req_path = rpath.Clean("/" + req_path)
	tmpv.writeView(req_path, h, w)
}

func (tmpv *TmplView) writeView(req_path string, r_header Getter, w HttpWriter) {
	w_header := w.Header()
	htreq, ht_err := htpath.New(tmpv.SystemFS, tmpv.DocumentRoot.String(),
		req_path, tmpv.IndexName)
	switch {
	case ht_err == nil:
	case ht_err == htpath.ErrInvalidDirRequest:
		if !tmpv.DirectoryRedirection {
			w.Error("404 invalid directory URL", http.StatusNotFound)
			return
		}

		p := rpath.SetDir(rpath.Join(tmpv.UrlTopPath, req_path))
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
	if dir_mod, ok := tmpv.DirViewStamp.DirModTime(htreq.Dir()); ok {
		htreq.UpdateModTime(dir_mod)
	}

	user := r_header.Get(tmpv.AuthnUserHeader)

	req_rpath := htreq.Req()
	is_dir := htreq.IsDir()
	has_doc := htreq.HasDoc()

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
	var fm_param *md2html.FrontMatterParam = &md2html.FrontMatterParam{}
	if has_doc {
		var rd_err error
		raw_bin, rd_err = unifs.ReadFile(tmpv.SystemFS, htreq.FullDoc())
		if rd_err != nil && !os.IsNotExist(rd_err) {
			w.Error("500 document file read error",
				http.StatusInternalServerError)
			return
		}

		if (proc_type == "md" ||
			proc_type == "text" && tmpv.CustomPageConfig.FrontMatter.UsedForText ||
			proc_type == "html" && tmpv.CustomPageConfig.FrontMatter.UsedForHtml) &&
			tmpv.CustomPageConfig.FrontMatter.IsEnabled() {

			body, fmp, fm_err := tmpv.CustomPageConfig.FrontMatter.TrimAndParse(raw_bin)
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

	theme_style := tmpv.ThemeStyle
	if fm_param.ThemeStyle != "" {
		theme_style = fm_param.ThemeStyle
	}

	page_style := tmpv.PageStyle
	if fm_param.PageStyle != "" {
		page_style = fm_param.PageStyle
	}
	style_tmpl := ""
	if page_style != "" {
		style_tmpl = "style_" + page_style + ".tmpl"
	}

	page_size_css := tmpv.PrintSizeCss
	if fm_param.PaperType != "" {
		css, ok := tmpv.PrintPaperMapping.GetCss(fm_param.PaperType)
		if ok {
			page_size_css = css
		}
	}

	print_zoom := tmpv.PrintZoom
	if fm_param.PrintZoom > 0 {
		print_zoom = fm_param.PrintZoom
	}

	loc_navi := tmpv.LocationNavi
	if fm_param.LocationNavi != "" {
		loc_navi = fm_param.LocationNavi
	}

	toc_navi := tmpv.TocNavi
	if fm_param.TocNavi != "" {
		toc_navi = fm_param.TocNavi
	}

	dir_view_mode := tmpv.DirectoryViewMode
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

	link_menu := []md2html.Link{}
	if tmpv.CustomPageConfig.LinkMenu.Default != nil {
		link_menu = tmpv.CustomPageConfig.LinkMenu.Default
	}
	if fm_param.LinkMenu != nil {
		link_menu = fm_param.LinkMenu
	}

	custom_param := md2html.CustomParam{}
	if tmpv.CustomPageConfig.CustomParam.Default != nil {
		custom_param = tmpv.CustomPageConfig.CustomParam.Default
	}
	if fm_param.CustomParam != nil {
		custom_param = fm_param.CustomParam
	}

	tmpl_funcs := template.FuncMap{
		"in_group": func(grp string) bool {
			return tmpv.UserMap.InGroup(user, grp)
		},
		"in_user": func() bool {
			return tmpv.UserMap.InUser(user)
		},
	}
	tmplext.AddDefaultFunc(tmpl_funcs, tmpv.SystemFS, tmpv.SvgIconPath)
	tmpl = tmpl.Funcs(tmpl_funcs)

	tmpl, err = tmpl.Parse(string(raw_bin))
	if err != nil {
		w.Error("503 service unavailable: "+err.Error(),
			http.StatusServiceUnavailable)
		return
	}

	var title_bin []byte

	title_bin = []byte("View: " + req_abs_path)

	with_title_param := false
	if fm_param.Title != "" {
		with_title_param = true
		title_bin = []byte(fm_param.Title)
	}

	var sm_card *md2html.SmCardParam = nil
	if tmpv.CustomPageConfig.SmCard.Enabled {
		sm_card = &md2html.SmCardParam{}
		*sm_card = fm_param.SmCard
		sm_card.Fix(&tmpv.CustomPageConfig.SmCard, req_abs_path)

		if sm_card.Title == "" {
			sm_card.Title = string(title_bin)
		}
	}

	tmpl_param := tmplHtmlParam{
		Options: &tmplOptions{
			ThemeStyle:    theme_style,
			PageStyle:     page_style,
			PrintSizeCss:  page_size_css,
			PrintZoom:     print_zoom,
			LocationNavi:  loc_navi,
			TocNavi:       toc_navi,
			DirectoryView: (dir_view_mode != "none"),
		},
		Markdown: tmpv.MarkdownConfig,
		SmCard:   sm_card,

		Title:     string(title_bin),
		Top:       tmpv.UrlTopPath,
		Lib:       tmpv.UrlLibPath,
		Path:      req_abs_path,
		PathLinks: links.NewLinks(rpath.Join("/", req_rpath)),
		LinkMenu:  link_menu,
		Files:     f_list,
		IsOpen:    is_open,

		UserName: user,

		CustomParam: custom_param,
	}

	var buf bytes.Buffer
	if e := tmpl.Execute(&buf, tmpl_param); e != nil {
		w.Error("500 html template execute error:"+e.Error(),
			http.StatusInternalServerError)
		return
	}

	var doc_bin []byte
	var toc_bin []byte

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
	case "text":
		doc_bin = buf.Bytes()
		toc_bin = []byte{}
	case "md":
		m2h := md2html.NewMd2Html(&md2html.Md2HtmlConfig{
			MdConfig:    tmpv.MarkdownConfig,
			SystemIds:   tmpv.SystemHtmlIds,
			SystemFS:    tmpv.SystemFS,
			FrontMatter: tmpv.CustomPageConfig.FrontMatter,
			StartMdFile: htreq.FullDoc(),
		})

		if fm_param.MarkdownConfig != "" {
			name := fm_param.MarkdownConfig
			if name[0] != '/' {
				name = rpath.Join(rpath.Dir(htreq.FullDoc()), name)
			}
			if strings.HasPrefix(name, tmpv.DocumentRoot.String()) {
				if md_cfg, err := md2html.NewMdConfig(tmpv.SystemFS, name); err == nil {
					m2h = m2h.NewLocalSpec(md_cfg)
				}
			}
		}

		var cerr error
		var md_title_bin []byte
		doc_bin, toc_bin, md_title_bin, cerr = m2h.Convert(buf.Bytes())
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

	mdtmpl_param := tmplMdParam{
		Options: &tmplOptions{
			ThemeStyle:    theme_style,
			PageStyle:     page_style,
			PrintSizeCss:  page_size_css,
			PrintZoom:     print_zoom,
			LocationNavi:  loc_navi,
			TocNavi:       toc_navi,
			DirectoryView: (dir_view_mode != "none"),
		},
		Markdown: tmpv.MarkdownConfig,
		SmCard:   sm_card,

		Title:     string(title_bin),
		Top:       tmpv.UrlTopPath,
		Lib:       tmpv.UrlLibPath,
		Path:      req_abs_path,
		PathLinks: links.NewLinks(rpath.Join("/", req_rpath)),
		LinkMenu:  link_menu,
		Files:     f_list,
		IsOpen:    is_open,

		Text:     string(doc_bin),
		TextType: text_type,
		Toc:      string(toc_bin),

		CustomParam: custom_param,
	}

	tmpl = tmplLookups(tmpl, style_tmpl, tmpv.MdTmplName)
	if tmpl == nil {
		w.Error("503 not found template", http.StatusServiceUnavailable)
		return
	}

	var mdbuf bytes.Buffer
	if e := tmpl.Execute(&mdbuf, mdtmpl_param); e != nil {
		w.Error("500 markdown template execute error:"+e.Error(),
			http.StatusInternalServerError)
		return
	}

	w_header.Set("Content-Type", "text/html; charset=UTF-8")
	w_header.Set("Last-Modified", last_mod)
	w_header.Set("Etag", tag)
	tmpv.setCacheHeader(w_header)
	mdbuf.WriteTo(w)
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

func (tmpv *TmplView) SumTemplate() (*TmplSum, error) {
	tmpl, err := tmpv.OriginTmpl.Clone()
	if err != nil {
		return nil, err
	}
	tmpl = tmpl.Lookup(tmpv.MdTmplName)
	if tmpl == nil {
		return nil, fmt.Errorf("template: no template %q", tmpv.MdTmplName)
	}

	tmpl_funcs := template.FuncMap{
		"in_group": func(grp string) bool {
			return true
		},
		"in_user": func() bool {
			return true
		},
	}
	tmplext.AddDefaultFunc(tmpl_funcs, tmpv.SystemFS, tmpv.SvgIconPath)
	tmpl = tmpl.Funcs(tmpl_funcs)

	tmpl_param := tmplMdParam{
		Options: &tmplOptions{
			ThemeStyle:    tmpv.ThemeStyle,
			PageStyle:     tmpv.PageStyle,
			PrintSizeCss:  tmpv.PrintSizeCss,
			PrintZoom:     tmpv.PrintZoom,
			LocationNavi:  tmpv.LocationNavi,
			TocNavi:       tmpv.TocNavi,
			DirectoryView: (tmpv.DirectoryViewMode != "none"),
		},
		Markdown:  tmpv.MarkdownConfig,
		Top:       tmpv.UrlTopPath,
		Lib:       tmpv.UrlLibPath,
		Path:      tmpv.UrlTopPath,
		PathLinks: links.NewLinks("/"),
		LinkMenu:  nil,
		Text:      "TEST text",
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

	if e := tmpv.WriteTestCatUi(h_ctx); e != nil {
		return nil, e
	}
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
	srv := &http.Server{Addr: tmpv.SocketPath, Handler: http.HandlerFunc(tmpv.Handler)}
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

	serr := srv.Serve(lstn)
	switch serr {
	default:
		return new_err("HTTP server error: %v.", serr)
	case nil:
	case http.ErrServerClosed:
	}

	return nil
}
