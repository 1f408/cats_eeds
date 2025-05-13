package alerts

import (
	"github.com/yuin/goldmark/ast"
)

type AlertBlockNode struct {
	ast.BaseBlock
	Label []byte
}

func (n *AlertBlockNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

var KindAlertBlock = ast.NewNodeKind("AlertBlock")

func (n *AlertBlockNode) Kind() ast.NodeKind {
	return KindAlertBlock
}

func NewAlertBlockNode(label []byte) *AlertBlockNode {
	n := &AlertBlockNode{
		BaseBlock: ast.BaseBlock{},
		Label: label,
	}
	n.SetAttributeString("class", "markdown-alert markdown-alert-"+string(label))
	return n
}

type AlertTitleNode struct {
	ast.BaseBlock
	Label []byte
}

func (n *AlertTitleNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
}

var KindAlertTitle = ast.NewNodeKind("AlertTitle")

func (n *AlertTitleNode) Kind() ast.NodeKind {
	return KindAlertTitle
}

func NewAlertTitleNode(label []byte) *AlertTitleNode {
	n := &AlertTitleNode{
		BaseBlock: ast.BaseBlock{},
		Label: label,
	}
	n.SetAttributeString("class", "markdown-alert_title markdown-alert_title-"+string(label))
	return n
}

