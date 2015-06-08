package app

import (
	goUrl "net/url"
	"path"
	"regexp"
	"strings"
)

const (
	ROUTER_METHOD_GET    = "GET"
	ROUTER_METHOD_POST   = "POST"
	ROUTER_METHOD_PUT    = "PUT"
	ROUTER_METHOD_DELETE = "DELETE"
)

type Route struct {
	regex  *regexp.Regexp
	method string
	params []string
	fn     []Handler
}

type Handler func(context *Context)

// router cache, save route param for caching.
type routerCache struct {
	param map[string]string
}

type Router struct {
	fn         []Handler
	routeSlice []*Route
}

func NewRouter() *Router {
	r := new(Router)
	r.routeSlice = make([]*Route, 0)
	return r
}

func newRoute() *Route {
	r := new(Route)
	r.params = make([]string, 0)
	return r
}

func (this *Router) Get(pattern string, fn ...Handler) {
	this.Router(ROUTER_METHOD_GET, pattern, fn...)
}

func (this *Router) Post(pattern string, fn ...Handler) {
	this.Router(ROUTER_METHOD_POST, pattern, fn...)
}

func (this *Router) Put(pattern string, fn ...Handler) {
	this.Router(ROUTER_METHOD_PUT, pattern, fn...)
}
func (this *Router) Delete(pattern string, fn ...Handler) {
	this.Router(ROUTER_METHOD_DELETE, pattern, fn...)
}

func (this *Router) Router(method, pattern string, fn ...Handler) {
	route := newRoute()
	route.method = method
	regex, params := this.parsePattern(pattern)
	route.params = params
	route.regex = regex
	route.fn = fn
	this.routeSlice = append(this.routeSlice, route)
}

func (this *Router) parsePattern(pattern string) (regx *regexp.Regexp, params []string) {
	params = make([]string, 0)
	segments := strings.Split(goUrl.QueryEscape(pattern), "%2F")
	for i, v := range segments {
		if strings.HasPrefix(v, "%3A") {
			segments[i] = `([\w-%]+)`
			params = append(params, strings.TrimPrefix(v, "%3A"))
		}
	}
	regx, _ = regexp.Compile("^" + strings.Join(segments, "/") + "$")
	return
}

func (this *Router) Find(url, method string) (params map[string]string, fn []Handler) {
	sfx := path.Ext(url)
	url = strings.Replace(url, sfx, "", -1)
	url = goUrl.QueryEscape(url)
	if !strings.HasSuffix(url, "%2F") && sfx == "" {
		url += "%2F"
	}
	url = strings.Replace(url, "%2F", "/", -1)
	for _, r := range this.routeSlice {
		if r.regex.MatchString(url) && r.method == method {
			p := r.regex.FindStringSubmatch(url)
			if len(p) != len(r.params)+1 {
				continue
			}
			params = make(map[string]string)
			for i, n := range r.params {
				params[n] = p[i+1]
			}
			fn = r.fn
			return
		}
	}
	return nil, nil
}
