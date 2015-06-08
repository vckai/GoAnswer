package app

import (
	"fmt"
	"net/http"
	"strings"
)

type App struct {
	config *ConfigFile
	inter  map[string]Handler
	view   *View
	router *Router
}

func NewApp() *App {
	a := new(App)
	a.config = LoadConfigFile("conf/app.conf")
	a.view = NewView("view")
	a.inter = make(map[string]Handler)
	a.router = NewRouter()
	return a
}

func (this *App) Config() *ConfigFile {
	return this.config
}

func (this *App) handler(res http.ResponseWriter, req *http.Request) {

	context := NewContext(this, res, req)

	if _, ok := this.inter["static"]; ok {
		this.inter["static"](context)
		if context.IsEnd {
			return
		}
	}
	if context.IsSend {
		return
	}
	var (
		params map[string]string
		fn     []Handler
		url    = req.URL.Path
	)
	params, fn = this.router.Find(url, req.Method)
	if params != nil && fn != nil {
		context.RouteParams = params

		for _, f := range fn {
			f(context)
			if context.IsEnd {
				break
			}
		}
		if !context.IsEnd {
			context.End()
		}
	} else {
		fmt.Println("router is missing at" + url)
		context.Status = 404
		if _, ok := this.inter["notfound"]; ok {
			this.inter["notfound"](context)
			if !context.IsEnd {
				context.End()
			}
		} else {
			context.Throw(404)
		}
	}
	context = nil
}

func (this *App) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	this.handler(res, req)
}

func (this *App) Run() {
	addr := this.config.MustValue("server", "localhost:9999")
	fmt.Println("http server run at " + addr)
	e := http.ListenAndServe(addr, this)
	fmt.Println("LISTEN SERVER ERROR: ", e)
}

func (this *App) GetString(key string) string {
	val, _ := this.config.GetValue("app." + key)
	return val
}

func (this *App) Get(method string, fn ...Handler) {
	this.router.Get(method, fn...)
}

func (this *App) Post(method string, fn ...Handler) {
	this.router.Post(method, fn...)
}

func (this *App) Put(method string, fn ...Handler) {
	this.router.Put(method, fn...)
}

func (this *App) Delete(method string, fn ...Handler) {
	this.router.Delete(method, fn...)
}

func (this *App) Route(method, key string, fn ...Handler) {
	methods := strings.Split(method, ",")
	for _, m := range methods {
		switch m {
		case "GET":
			this.Get(key, fn...)
		case "POST":
			this.Post(key, fn...)
		case "PUT":
			this.Put(key, fn...)
		case "DELETE":
			this.Delete(key, fn...)
		default:
			fmt.Println("unknow route method " + m)
		}
	}
}

func (this *App) Static(h Handler) {
	this.inter["static"] = h
}

func (this *App) NotFound(h Handler) {
	this.inter["notfound"] = h
}
