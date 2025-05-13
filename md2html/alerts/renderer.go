package alerts

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

var AlertBlockAttributeFilter = html.GlobalAttributeFilter

type alertBlockRenderer struct {
	Config
}

func NewAlertBlockRenderer(opts ...Option) renderer.NodeRenderer {
	r := &alertBlockRenderer {
		Config: DefaultConfig,
	}
	for _, opt := range opts {
		opt.SetAlertBlockOption(&r.Config)
	}
	return r
}

func (r *alertBlockRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindAlertBlock, r.renderAlertBlock)
}

func (r *alertBlockRenderer) renderAlertBlock(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*AlertBlockNode)
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<div")
			html.RenderAttributes(w, n, AlertBlockAttributeFilter)
			_, _ = w.WriteString(">\n")
		} else {
			_, _ = w.WriteString("<div>\n")
		}
	} else {
		_, _ = w.WriteString("</div>\n")
	}
	return ast.WalkContinue, nil
}

var AlertTitleAttributeFilter = html.GlobalAttributeFilter

type alertTitleRenderer struct {
	Config
}

func NewAlertTitleRenderer(opts ...Option) renderer.NodeRenderer {
	r := &alertTitleRenderer {
		Config: DefaultConfig,
	}
	for _, opt := range opts {
		opt.SetAlertBlockOption(&r.Config)
	}
	return r
}

func (r *alertTitleRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindAlertTitle, r.renderAlertTitle)
}

func (r *alertTitleRenderer) renderAlertTitle(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*AlertTitleNode)
	if entering {
		if n.Attributes() != nil {
			_, _ = w.WriteString("<p")
			html.RenderAttributes(w, n, AlertTitleAttributeFilter)
			_ = w.WriteByte('>')
		} else {
			_, _ = w.WriteString("<p>")
		}
		if hc, ok := r.Config.TitleHtmlMapping[string(n.Label)]; ok {
			_, _ = w.WriteString(hc)
		}
	} else {
		_, _ = w.WriteString("</p>\n")
	}
	return ast.WalkContinue, nil
}
