package http_router

import (
	"context"
	"net/http"
	"sync"
)

// Handle is a function that can be registered to a route to handle HTTP request
// Like http.HandleFunc, but has a third parameter for the values of wildcards
type Handle func(http.ResponseWriter, *http.Request, Params)

// Param is a single URL parameter,consisting of a key and a value
type Param struct {
	Key   string
	Value string
}

// Params is a Param-slice, as returned by the router
type Params []Param

// ByName returns the value of the first Param which key matches the given name
func (ps Params) ByName(name string) string {
	for _, p := range ps {
		if p.Key == name {
			return p.Value
		}
	}
	return ""
}

type paramsKey struct{}

// ParamKey is the request context key under which URL params are stored
var ParamsKey = paramsKey{}

// ParamsFromContext pulls the URL parameters from a request context
// or return nil if none are present
func ParamsFromContext(ctx context.Context) Params {
	p, _ := ctx.Value(ParamsKey).(Params)
	return p
}

// Router is a http.Handler which can be used to dispatch request to different
// handler functions via configurable routes
type Router struct {
	paramsPool sync.Pool
	maxParams  uint16

	// 开启后在调用handler之前将匹配的路由添加到http.Request上下文
	SaveMatchedRoutePath bool

	// 无法匹配是否启动重定向
	RedirectTrailingSlash bool

	// 开启后会修复当前的请求路由
	RedirectFixedPath bool

	// 开启后会检查是否允许使用另一种方法
	HandleMethodNotAllowed bool

	// 开启后自动回复options 请求
	HandleOPTIONS bool

	// Cached value of global (*) allowed methods
	globalAllowed string

	// option 请求的处理handle
	GlobalOPTIONS http.Handler

	// 无路由匹配的handle
	NotFound http.Handler

	// 方法不匹配的handle
	MethodNotAllowed http.Handler

	PanicHandler func(http.ResponseWriter, *http.Request, interface{})
}

func New() *Router {
	return &Router{
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      true,
		HandleMethodNotAllowed: true,
		HandleOPTIONS:          true,
	}
}
