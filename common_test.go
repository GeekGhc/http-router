package http_router

import "testing"

func TestFindWildcard(t *testing.T) {
	path1 := "/home/:name/info"
	// name 6 true
	wildcard, index, valid := findWildcard(path1)
	t.Log("wildcard:", wildcard, " index:", index, " valid:", valid)

	path2 := "/home/:name:/info"
	// :name: 6 false
	wildcard, index, valid = findWildcard(path2)
	t.Log("wildcard:", wildcard, " index:", index, " valid:", valid)
}

func TestShiftNRuneBytes(t *testing.T) {
	bytes := [4]byte{1, 1, 1, 1}
	res := shiftNRuneBytes(bytes, 1)
	t.Log(res)
}
