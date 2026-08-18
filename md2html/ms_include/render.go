package ms_include

import (
	"bytes"
	"path"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

type IncludeHTMLRenderer struct {
	convertHtml ConvertHtmlFunc
	pathStack   PathStack
}

func NewIncludeHTMLRenderer(conv ConvertHtmlFunc, ps PathStack) renderer.NodeRenderer {
	return &IncludeHTMLRenderer{
		convertHtml: conv,
		pathStack:   ps,
	}
}

func (r *IncludeHTMLRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindInclude, r.renderInclude)
}

func (r *IncludeHTMLRenderer) renderInclude(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	includeNode := node.(*IncludeNode)
	file := string(includeNode.Link)
	if !isFilePath(file) {
		return ast.WalkContinue, nil
	}

	file = path.Clean(file)
	if len(file) <= 0 {
		return ast.WalkContinue, nil
	}

	cwd := r.pathStack.Cwd()

	if file[0] != '/' {
		file = path.Join(cwd, file)
	}

	if err := r.pathStack.Push(file); err != nil {
		return ast.WalkContinue, nil
	}

	var buf bytes.Buffer
	if e := r.convertHtml(file, &buf); e != nil {
		return ast.WalkContinue, nil
	}
	buf.WriteTo(w)

	r.pathStack.Pop()

	return ast.WalkSkipChildren, nil
}
