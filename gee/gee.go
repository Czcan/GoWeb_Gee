package gee

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

type HandlerFunc func(*Context)

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc //	support middleware
	engine      *Engine       //	all group share one engine instance
}

type Engine struct {
	*RouterGroup // go的嵌套类型，这样Engine就有RouterGroup的属性了， 相当于java或python等语言的 继承， Engine 继承于 RouterGroup，子类Engine比父类有更多的成员变量和属性

	// 路由映射表
	router *router
	groups []*RouterGroup // store all RouterGroup

	htmlTemplates *template.Template // for html render
	funcMap       template.FuncMap   // for html render
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
		engine: engine,
	}

	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

// User is define to add middleware to the group
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
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

// create static handler
func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	absolutPath := path.Join(group.prefix, relativePath)
	fileServer := http.StripPrefix(absolutPath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		// Check the file is existed and/or is have the permission to access it
		if _, err := os.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}

		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// serve static file
// 用户 可以将 磁盘目录上的 某个文件夹root 映射 到 relativepath
// 比如 Static("/assets", "/user/local/htwer") 用户访问localhost:9999/assets/js/css 最终返回 /user/local/htwer/js/css
func (group *RouterGroup) Static(relativePath string, root string) {
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// Register Get handler
	group.GET(urlPattern, handler)
}

// for custom render funcion
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.funcMap = funcMap
}

func (engine *Engine) LoadHTMLGlob(pattern string) {
	engine.htmlTemplates = template.Must(template.New("").Funcs(engine.funcMap).ParseGlob(pattern))
}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

// 解析请求的路径， 查找路由映射表 -> 查到则执行注册的处理方法，否则， 返回错误码
func (engine *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var middlewares []HandlerFunc

	for _, group := range engine.groups {
		if strings.HasPrefix(r.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}

	c := newContext(w, r)
	c.handlers = middlewares
	c.engine = engine
	engine.router.handle(c)
}
