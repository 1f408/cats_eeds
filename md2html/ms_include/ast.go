package ms_include

import (
	"github.com/yuin/goldmark/ast"
)

var KindInclude = ast.NewNodeKind("Include")

type IncludeNode struct {
	ast.BaseInline
	Destination []byte
	Title []byte
	Link []byte
}

func (n *IncludeNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

func (n *IncludeNode) Kind() ast.NodeKind {
    return KindInclude
}

func NewIncludeNode() *IncludeNode {
    return &IncludeNode{
        BaseInline: ast.BaseInline{},
    }
}
