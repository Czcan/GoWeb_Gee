# GoWeb_Gee

框架：

1. 路由：将 请求 映射到 函数， 支持动态路由，如 '/hello/:name'
2. 模板：内置模板提供 模板渲染 
3. 工具集： 提供对 cookies， headers 等的 处理机制
4. 插件(中间件)： 全局 / 特定路由



性能是 Router(路由) 的 重要指标之一  -> 考虑 匹配算法 用 Trie树 实现

### HTTP基础

我们需要抽象出 框架的引擎 Engine, 为我们提供一些基础的方法， 以及更新注册路由等方法

我们需要启动服务， 所以需要实现 ServeHTTP方法， 剩下的是一些 路由相关的实现

路由的参数必须是两个，一个 http.ResponseWriter,  向客户端回写数据，  一个*http.Request,  接受客户端请求 （第二个参数用指针类型是因为Request是结构体，用指针节省内存， 而Writer是个接口类型， 不能用指针）



这一部分 **基于net/http库**进行二次封装，

实现Engine之前，用http.HandlerFunc实现路由和handler的映射，只能针对具体路由写处理逻辑 

**统一的控制入口** ->  实现Engine后，所有HTTP请求都被拦截 ,  我们可以在Engine中自定义路由映射的逻辑，也可以统一添加处理逻辑，如日志，错误处理等





实现http.Handler接口（type Handler interface { ServeHTTP(w, r)}） 

// go.mod 中 使用 replace 让 gee 指向 ./gee

->  gee.go (Engine Struct,  New() , addRoute(),  GET(), POST(),  ServeHTTP(),  Run())

// 私有包如果不想发布到网上，需要手动添加require ，然后replace 进行替换，将私有包指向本地module所在的绝对或相对路径。一般用相对路径更通用



1. Run()中 调用 http.ListenServer(addr, **engine**);  

// 第二个参数是Handler类型,是个接口类型,所以只要是实现了Handler中的方法(即ServeHTTP)，那么就相当于是Handler类型了，这里传的是engine，用engine来处理所有的HTTP请求，

//第二个参数传的nil -> 标准库用默认的DefaultMux这个接口

http.ListenServe() 底层自动调用了 ServeHTTP， 所以 engine重写ServeHTTP后 不用显式调用

2. Engine中的 路由映射表 router map[string]HandlerFunc  ， key用 请求方法method + “-” + pattern  ->  针对相同的路由，如果请求方法不同，可以映射不同的处理方法
3. HandlerFunc()  是给框架用户自定义路由映射方法



这部分实现了路由映射表，提供了用户注册静态路由的方法， 包装了启动函数



### 上下文 Context

1. 将router独立，方便后续增强、
2. 封装Request，ResponseWriter，提供对JSON，HTML， String等返回类型的支持

**上下文Context承载了与当前请求强相关的信息，路由处理函数，中间件等的 参数， 就像一次会话的百宝箱，可以找到任何东西**

**扩展性和复杂性留在内部，对外简化接口**



设置Header正确调用顺序Header().Set 然后WriteHeader() 最后是Write(), 如果先WriterHeader()， 那么Header.Set()不会生效



### 前缀树路由

使用 Trie树 实现 动态路由(Dynamic router)解析

动态路由有很多实现方式，支持的 规则，性能 有很大差异， 如开源的gorouter，支持在路由规则中嵌入正则表达式，即路径中的参数仅支持数字或者字母。 开源的httprouter则不支持正则表达式



前缀树（Trie树）：每一个节点 的 子节点 都拥有 相同的前缀

HTTP路径恰好有 / 分隔的多段构成  

->  每一段 作为 前缀树 的 一个节点

->  通过树结构查询， 如果 中间某一层的节点 都不满足条件， 那么意味着 没有匹配的路由， 查询结束



Trie树 节点 node {

pattern string;  // 待匹配路由

part string;  // 路由一部分

children []*node;  // 子节点

isWild bool;  // 是否模糊匹配

}

为了实现动态路由  ->  Trie树节点 增加 isWild参数， 表示 是否模糊匹配， 即判断节点part， part[0] == ":" || part[0] == "*"



路由 最重要的两点  注册， 匹配   ->  Trie树 需要支持 节点的 插入 ， 查询

插入功能： 递归查找每一层的节点， 如果没有匹配到part的节点， 就新建一个， 注意只有路由最后一个节点的pattern才存路径，其他节点的pattern都为空 （如 /lang/:name/doc, 只有doc.pattern == "/lang/:name/doc", lang.pattern, :name.pattern都 == ""） -》 当匹配结束时，可以通过 n.pattern == "" 来判断 是否匹配成功

查询功能：递归每一层节点， 退出规则是 匹配到* / 匹配失败 / 匹配到了 第len(parts)层 节点



Router

Trie树应用到router， roots来存每种请求方式的Trie树根节点，  handlers 来存 每种请求方式的 HandlerFunc

getRoute方式中 还 解析 了两种匹配符“ * ” ， “ : ” 的参数， 返回一个map。 

```go
如 /p/go/doc 匹配到了 /p/:lang/doc,那么返回{lang:"go"}
/static/css/geektutu.css`匹配到`/static/*filepath`，解析结果为`{filepath: "css/geektutu.css"}
```



Context与handle的变化

HandlerFunc中希望能够访问到解析的参数， 所以 Context中多加一个属性map[string]string 参数 和 一个方法 c.Param() 来提供对路由参数的访问

解析后的参数存到Params中， 通过如Param("lang") 来访问



在调用匹配到的`handler`前，将解析出来的路由参数赋值给了`c.Params`。这样就能够在`handler`中，通过`Context`对象访问到具体的值了


### 分组控制

中间件可以给框架无限的扩展能力， 应用在分组上，可以让分组控制的收益更加明显， 而不单单只是共享相同的前缀而已。

比如“/admin”的分组，可以应用鉴权中间件，“ / ”的分组可以应用日志中间件，这意味着全局路由都应用日志中间件。

1. 为什么Engine要RouterGroup类型？

   进一步抽象，将Engine作为最顶层的分组，将路由相关的函数交给RouterGroup实现，而Engine拥有RouterGroup的所有能力。	RouterGroup类型控制路由相关的功能，而Engine作为框架的引擎， 不单单只是路由相关，整个框架的所有资源都是由Enigne统一协调，这样就可以通过Engine间接访问所有接口。

    go中的嵌套类型，相当于java/python的继承，Engine中嵌套了RouterGroup类型， 这样Engine就拥有了RouterGroup 的所有属性和方法、

2. 为什么RouterGroup里要保存一个Engine指针

因为我们想要 Group对象可以直接映射路由规则， 如 v1 := r.Group("v1");  v1.GET(....),而Engine是框架的引擎，所有资源都由Engine统一协调，我们可以通过Engine间接访问所有接口，即Group对象可以通过Engine访问到router。

保存的是指针，所以所有RouterGroup对象访问的是同一个Engine实例zz

3. 前缀分组路由是一种比较好的方式

如对一个博客系统来说， /auth,  /user走授权中间件（使用者相关，需要授权）， /post/:id/html（网页自身统计一些信息触发，不需要授权）就不走授权，走统计中间件， /api 开头的 是对外的公共接口，诸如此类，是比较符合URL设计的习惯的。

特别是Restful API， 以资源为中心的URL设计， 通过前缀做不同的业务的区分更为明显，不同前缀代表不同类型的资源


### 中间件

1. 为什么Context中的的req要指针类型， 而writer不用指针类型

Requset是结构体类型，用指针可以节省内存（因为go里变量是值类型的，结构体变量赋给另一个，是一次内存拷贝），而Writer是接口类型，没必要用指针（接口类型本身就是引用类型）。

1. index初始化-1的原因

c.Next()一开始调用了c.index++， 这样刚好从0开始

2. Next()函数需要遍历 handlers的原因

执行到某个中间件A中的c.Next() 后， 先执行它后面的中间件，执行完后，才继续执行中间件A中 c.Next()后面的内容

3. 如何实现 类似 gin 的 c.Abort() 这种 中间件 退出机制

在Context中维护一个状态值，调用c.Abort()改变状态，循环时检查状态，发现已经中止停止循环。