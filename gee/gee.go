package gee

import (
	"log"
	"net/http"
)

type HandlerFunc func(c *Context)

type Engine struct {
	*RouterGroup // go的嵌套类型，这样Engine就有RouterGroup的属性了， 相当于java或python等语言的 继承， Engine 继承于 RouterGroup，子类Engine比父类有更多的成员变量和属性

	// 路由映射表
	router *router
	groups []*RouterGroup // store all RouterGroup
}

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc //	support middleware
	parent      *RouterGroup
	engine      *Engine //	all group share a engine instance
}

func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

// Group is define to create a new RouterGroup
// Remeber all RouterGroup share the same Engine instance
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}

	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route%4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
}

func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	// group.engine.addRoute(...)
	group.addRoute("GET", pattern, handler) // 注意这里addRoute已经是绑定在group,所以直接group.addRoute()...
}

func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

// 解析请求的路径， 查找路由映射表 -> 查到则执行注册的处理方法，否则， 返回错误码
func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c := newContext(w, r)
	engine.router.handle(c)
}
