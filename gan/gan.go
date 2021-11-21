package gan

import (
	"net/http"
	"strings"
)

type HandlerFunc func(c *Context)

type Engine struct {
	*RouterGroup
	router *router
	groups []*RouterGroup
}

func New() *Engine {
	e := &Engine{router: newRouter()}
	e.RouterGroup = &RouterGroup{engine: e}
	e.groups = []*RouterGroup{e.RouterGroup}
	return e
}

func Default() *Engine {
	e := New()
	e.Use(Logger(), Recovery())
	return e
}

func (e *Engine) addRoute(method, pattern string, handler HandlerFunc) {
	e.router.addRoute(method, pattern, handler)
}

func (e *Engine) GET(pattern string, handler HandlerFunc) {
	e.addRoute("GET", pattern, handler)
}

func (e *Engine) POST(pattern string, handler HandlerFunc) {
	e.addRoute("POST", pattern, handler)
}

func (e *Engine) Run(addr string) error {
	return http.ListenAndServe(addr, e)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var middlewares []HandlerFunc
	for _, g := range e.groups {
		if strings.HasPrefix(r.URL.Path, g.prefix) {
			middlewares = append(middlewares, g.middlewares...)
		}
	}
	c := newContext(w, r)
	c.handlers = middlewares
	e.router.handle(c)
}

type RouterGroup struct {
	parent      *RouterGroup
	prefix      string
	middlewares []HandlerFunc
	engine      *Engine
}

func (g *RouterGroup) Group(prefix string) *RouterGroup {
	e := g.engine
	newG := &RouterGroup{
		parent: g,
		prefix: g.prefix + prefix,
		engine: e,
	}
	e.groups = append(e.groups, newG)
	return newG
}

func (g *RouterGroup) addRoute(method, pattern string, handler HandlerFunc) {
	g.engine.addRoute(method, g.prefix+pattern, handler)
}

func (g *RouterGroup) GET(pattern string, handler HandlerFunc) {
	g.addRoute("GET", pattern, handler)
}

func (g *RouterGroup) POST(pattern string, handler HandlerFunc) {
	g.addRoute("POST", pattern, handler)
}

func (g *RouterGroup) Use(middlewares ...HandlerFunc) {
	g.middlewares = append(g.middlewares, middlewares...)
}
