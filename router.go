package http_router

import (
	"context"
	"net/http"
	"sync"
)

// 路由处理func,包含路由第三方参数
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

// 路由匹配后的参数名称
var MatchedRoutePathParam = "$matchedRoutePath"

func (ps Params) MatchedRoutePath() string {
	return ps.ByName(MatchedRoutePathParam)
}

// Router is a http.Handler which can be used to dispatch request to different
// handler functions via configurable routes
type Router struct {
	trees map[string]*node

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

// get路由参数
func (r *Router) getParams() *Params {
	ps, _ := r.paramsPool.Get().(*Params)
	*ps = (*ps)[0:0] //reset slice
	return ps
}

// set路由参数
func (r *Router) putParams(ps *Params) {
	if ps != nil {
		r.paramsPool.Put(ps)
	}
}

func (r *Router) saveMatchedRoutePath(path string, handle Handle) Handle {
	return func(w http.ResponseWriter, req *http.Request, ps Params) {
		if ps == nil {
			psp := r.getParams()
			ps = (*psp)[0:1]
			ps[0] = Param{Key: MatchedRoutePathParam, Value: path}
			handle(w, req, ps)
			r.putParams(psp)
		} else {
			ps = append(ps, Param{Key: MatchedRoutePathParam, Value: path})
			handle(w, req, ps)
		}
	}
}

func (r *Router) HEAD(path string, handle Handle) {
	r.Handle(http.MethodHead, path, handle)
}
func (r *Router) OPTIONS(path string, handle Handle) {
	r.Handle(http.MethodOptions, path, handle)
}
func (r *Router) GET(path string, handle Handle) {
	r.Handle(http.MethodGet, path, handle)
}
func (r *Router) POST(path string, handle Handle) {
	r.Handle(http.MethodPost, path, handle)
}
func (r *Router) PUT(path string, handle Handle) {
	r.Handle(http.MethodPut, path, handle)
}
func (r *Router) PATCH(path string, handle Handle) {
	r.Handle(http.MethodPatch, path, handle)
}
func (r *Router) DELETE(path string, handle Handle) {
	r.Handle(http.MethodDelete, path, handle)
}

func (r *Router) Handle(method, path string, handle Handle) {
	varsCount := uint16(0)

	// 基本校验
	if method == "" {
		panic("method must not be empty")
	}
	if len(path) < 1 || path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}
	if handle == nil {
		panic("handle must not be nil")
	}
	// 保存路由信息
	if r.SaveMatchedRoutePath {
		varsCount++
		handle = r.saveMatchedRoutePath(path, handle)
	}
	if r.trees == nil {
		r.trees = make(map[string]*node)
	}

	root := r.trees[method]
	if root == nil {
		root = new(node)
		r.trees[method] = root

		//r.globalAllowed = r
	}

	// update maxParams
	if paramsCount := countParams(path); paramsCount+varsCount > r.maxParams {
		r.maxParams = paramsCount + varsCount
	}

	// Lazy-Init paramsPool alloc func
	if r.paramsPool.New == nil && r.maxParams > 0 {
		r.paramsPool.New = func() interface{} {
			ps := make(Params, 0, r.maxParams)
			return &ps
		}
	}
}

func (r *Router) Handler(method, path string, handler http.Handler) {
	r.Handle(method, path,
		func(w http.ResponseWriter, req *http.Request, p Params) {
			if len(p) > 0 {
				ctx := req.Context()
				ctx = context.WithValue(ctx, ParamsKey, p)
				req = req.WithContext(ctx)
			}
			handler.ServeHTTP(w, req)
		},
	)
}
