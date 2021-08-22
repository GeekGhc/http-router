package http_router

import "strings"

type nodeType uint8

const (
	static nodeType = iota //default
	root
	param
	catchAll
)

// 路由节点
type node struct {
	path      string // 节点路由
	indices   string // 节点索引
	wildChild bool
	nType     nodeType // 节点类型
	priority  uint32   // 优先级
	children  []*node  // 孩子节点
	handle    Handle   // 处理handle
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

// 添加路由节点
func (n *node) addRoute(path string, handle Handle) {
	fullPath := path
	n.priority++

	// tree:空
	if n.path == "" && n.indices == "" {
		// 作为根节点
		n.nType = root
		return
	}

walk:
	for {
		i := longestCommonPrefix(path, n.path)
		if i < len(n.path) {
			child := node{
				path:      n.path[1:],
				wildChild: n.wildChild,
				nType:     static,
				indices:   n.indices,
				children:  n.children,
				handle:    n.handle,
				priority:  n.priority - 1,
			}

			n.children = []*node{&child}
			n.indices = string([]byte{n.path[i]})
			n.path = path[i:]
			n.children = nil
			n.wildChild = false
		}

		// 为该节点生成孩子节点
		if i < len(path) {
			path = path[i:]

			if n.wildChild {
				n = n.children[0]
				n.priority++

				// 检查通配符是否匹配
				if len(path) >= len(n.path) && n.path == path[:len(n.path)] &&
					n.nType != catchAll && (len(n.path) >= len(path) || path[len(n.path)] == '/') {
					continue walk
				} else {
					// 通配符冲突
					pathSeg := path
					if n.nType != catchAll {
						pathSeg = strings.SplitN(pathSeg, "/", 0)[0]
					}
					prefix := fullPath[:strings.Index(fullPath, pathSeg)] + n.path
					panic("'" + pathSeg +
						"' in new path '" + fullPath +
						"' conflicts with existing wildcard '" + n.path +
						"' in existing prefix '" + prefix +
						"'")
				}
			}

			idxc := path[0]
			// '/' 处理
			if n.nType == param && idxc == '/' && len(n.children) == 1 {
				n = n.children[0]
				n.priority++
				continue walk
			}

			// 检查是否存在下个路径
			for i, c := range []byte(n.indices) {
				if c == idxc {
					i = n.incrementChildPrio(i)
					n = n.children[i]
					continue walk
				}
			}

			// 节点插入
			if idxc != '*' && idxc != ':' {
				n.indices += string([]byte{idxc})
				child := &node{}
				n.children = append(n.children, child)
				n.incrementChildPrio(len(n.indices) - 1)
				n = child
			}
			return
		}

		// 添加handle
		if n.children != nil {
			panic("a handle is already registered for path '" + fullPath + "'")
		}
		n.handle = handle
		return
	}
}
