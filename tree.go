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
