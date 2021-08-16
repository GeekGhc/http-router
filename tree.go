package http_router

type nodeType uint8

const (
	static nodeType = iota //default
	root
	param
	catchAll
)

// 路由节点
type node struct {
	path      string
	indices   string
	wildChild bool
	nType     nodeType
	priority  uint32
	children  []*node
	handle    Handle
}

// 调整孩子节点优先级
func (n *node) incrementChildPrio(pos int) int {
	cs := n.children
	cs[pos].priority++
	prio := cs[pos].priority

	newPos := pos
	for ; newPos > 0 && cs[newPos-1].priority < prio; newPos-- {
		// 交换节点
		cs[newPos-1], cs[newPos] = cs[newPos], cs[newPos-1]
	}

	//生成新的index
	if newPos != pos {
		n.indices = n.indices[:newPos] +
			n.indices[pos:pos+1] +
			n.indices[newPos:pos] + n.indices[pos+1:]
	}
	return newPos
}
