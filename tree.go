package djson

type (
	nodeType  int
	nodeScope int
)

const (
	nodeRoot = nodeType(iota)
	nodeVariable
	nodeValue
	nodeArray
	nodeTemplate
	nodeMap
	nodeNested
	nodeUnknown
)

const (
	scopeGlobal = nodeScope(iota)
	scopeMap
	scopeArray
	scopeTemplate
	scopeNested
	scopeUnknown
	scopeCompare
)

type node struct {
	typ      nodeType
	scope    nodeScope
	children []*node
	parent   *node
	token    *Token
}

func (n *node) add(c *node) *node {
	n.children = append(n.children, c)
	c.parent = n
	return c
}

func (n *node) pop() *node {
	if len(n.children) == 0 {
		return nil
	}
	l := len(n.children)
	r := n.children[l-1]
	n.children = n.children[:l-1]
	r.parent = nil
	return r
}

type tree struct {
	root *node
	curr *node
}

func newTree() *tree {
	node := &node{
		typ:   nodeRoot,
		scope: scopeGlobal,
	}
	return &tree{root: node, curr: node}
}

func (t *tree) add(n *node) *node {
	t.curr.add(n)
	t.curr = n
	return n
}
