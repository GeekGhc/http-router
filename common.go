package http_router

// 统计':','*'个数
func countParams(path string) uint16 {
	var n uint
	for i := range []byte(path) {
		switch path[i] {
		case ':', '*':
			n++
		}
	}
	return uint16(n)
}

// 返回最小值
func minElem(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

// 最长匹配前缀
func longestCommonPrefix(a, b string) int {
	var i int
	max := minElem(len(a), len(b))
	for i < max && a[i] == b[i] {
		i++
	}
	return i
}

// 搜索通配符并检查非法字符
func findWildcard(path string) (wildcard string, i int, valid bool) {
	for start, c := range []byte(path) {
		// 匹配 ':' | '*'
		if c != ':' && c != '*' {
			continue
		}

		valid := true
		for end, c := range []byte(path[start+1:]) {
			switch c {
			case '/':
				return path[start : start+1+end], start, valid
			case ':', '*':
				valid = false
			}
		}
		return path[start:], start, valid
	}
	return "", -1, false
}

// 字节数组向左移动
func shiftNRuneBytes(rb [4]byte, n int) [4]byte {
	switch n {
	case 0:
		return rb
	case 1:
		return [4]byte{rb[1], rb[2], rb[3]}
	case 2:
		return [4]byte{rb[2], rb[3]}
	case 3:
		return [4]byte{rb[3]}
	default:
		return [4]byte{}
	}
}
