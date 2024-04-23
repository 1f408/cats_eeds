package tmplview

import (
	"fmt"
	"io"
	"net/http"
	"io/fs"
	"path"

	"github.com/l4go/unifs"
)

type Getter interface {
	Get(string) string
}

type DummyGetter struct{}

func (DummyGetter) Get(string) string {
	return ""
}

type Setter interface {
	Set(string, string)
}

type DummySetter struct{}

func (DummySetter) Set(string, string) {}

type HttpWriter interface {
	Header() Setter
	Write([]byte) (int, error)
	Error(string, int)
	WriteHeader(int)
	ServeFile(fsys fs.FS, file string)
}

type HttpWrite struct {
	w http.ResponseWriter
	r *http.Request
}

func NewHttpWriter(w http.ResponseWriter, r *http.Request) *HttpWrite {
	return &HttpWrite{w: w, r: r}
}

func (hw *HttpWrite) Header() Setter {
	return hw.w.Header()
}
func (hw *HttpWrite) Write(buf []byte) (int, error) {
	return hw.w.Write(buf)
}
func (hw *HttpWrite) WriteHeader(code int) {
	hw.w.WriteHeader(code)
}
func (hw *HttpWrite) Error(msg string, code int) {
	http.Error(hw.w, msg, code)
}

func (hw *HttpWrite) ServeFile(fsys fs.FS, file string) {
	name := path.Base(file)
	if name == "" || name == "." || name == "/" {
		hw.Error("404 not found", http.StatusNotFound)
	}

	f, err := unifs.Open(fsys, file)
	if err != nil {
		hw.Error("404 not found", http.StatusNotFound)
		return
	}
	fi, err := f.Stat()
	if err != nil {
		hw.Error("404 not found", http.StatusNotFound)
		return
	}

	if s, ok := f.(io.ReadSeeker); ok {
		http.ServeContent(hw.w, hw.r, name, fi.ModTime(), s)
		return
	}

	io.Copy(hw.w, f)
}

type DumpWrite struct {
	out  io.Writer
	eout io.Writer
}

func NewDumpWrite(out io.Writer, eout io.Writer) *DumpWrite {
	return &DumpWrite{out: out, eout: eout}
}

func (w *DumpWrite) Header() Setter {
	return &DummySetter{}
}
func (w *DumpWrite) Write(buf []byte) (int, error) {
	return w.out.Write(buf)
}

func (w *DumpWrite) WriteHeader(code int) {
	fmt.Fprintf(w.eout, "Code: %d\n", code)
	panic("must not be called")
}

func (w *DumpWrite) Error(msg string, code int) {
	fmt.Fprintf(w.eout, "Error: %d: %s\n", code, msg)
}

func (hw *DumpWrite) ServeFile(fsys fs.FS, file string){
	name := path.Base(file)
	if name == "" || name == "." || name == "/" {
		hw.Error("404 not found", http.StatusNotFound)
	}

	f, err := unifs.Open(fsys, file)
	if err != nil {
		hw.Error("404 not found", http.StatusNotFound)
		return
	}

	io.Copy(hw.out, f)
}
