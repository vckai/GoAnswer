package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	CONTEXT_RENDERED = "context_rendered"
	CONTEXT_END      = "context_end"
	CONTEXT_SEND     = "context_send"
)

type Context struct {
	Request    *http.Request
	BaseUrl    string
	Url        string
	RequestUrl string
	Method     string
	Ip         string
	UserAgent  string
	Referer    string
	Host       string
	Ext        string
	IsSSH      bool
	IsAjax     bool

	Response http.ResponseWriter
	Status   int
	Header   map[string]string
	Body     []byte

	RouteParams map[string]string

	eventsFunc map[string][]reflect.Value

	IsSend bool
	IsEnd  bool

	app *App
}

func NewContext(app *App, res http.ResponseWriter, req *http.Request) *Context {
	c := new(Context)
	c.app = app
	c.IsSend = false
	c.IsEnd = false

	c.Request = req
	c.Url = req.URL.Path
	c.RequestUrl = req.RequestURI
	c.Method = req.Method
	c.Host = req.Host
	c.Ip = strings.Split(req.RemoteAddr, ":")[0]
	c.IsAjax = req.Header.Get("X-Requested-With") == "XMLHttpRequest"
	c.IsSSH = req.TLS == nil
	c.UserAgent = req.UserAgent()
	c.Referer = req.Referer()

	c.eventsFunc = make(map[string][]reflect.Value)

	c.BaseUrl = "://" + c.Host + "/"
	if c.IsSSH {
		c.BaseUrl = "https" + c.BaseUrl
	} else {
		c.BaseUrl = "http" + c.BaseUrl
	}

	c.Response = res
	c.Status = 200
	c.Header = make(map[string]string)
	c.Header["Content-Type"] = "text/html;charset=UTF-8"

	req.ParseForm()
	return c
}

func (this *Context) Strings(key string) []string {
	return this.Request.Form[key]
}

func (this *Context) String(key string) string {
	return this.Request.FormValue(key)
}

func (this *Context) MustString(key, def string) string {
	value := this.String(key)
	if value == "" {
		return def
	}
	return value
}

// On registers event function to event name string.
func (this *Context) On(e string, fn interface{}) {
	if reflect.TypeOf(fn).Kind() != reflect.Func {
		println("only support function type for Context.On method")
		return
	}
	if this.eventsFunc[e] == nil {
		this.eventsFunc[e] = make([]reflect.Value, 0)
	}
	this.eventsFunc[e] = append(this.eventsFunc[e], reflect.ValueOf(fn))
}

func (this *Context) Do(e string, args ...interface{}) [][]interface{} {
	fns, ok := this.eventsFunc[e]
	if !ok {
		return nil
	}
	if len(fns) < 1 {
		return nil
	}
	resSlice := make([][]interface{}, 0)
	for _, fn := range fns {
		if !fn.IsValid() {
			fmt.Println("invalid event function caller for " + e)
		}
		numIn := fn.Type().NumIn()
		if numIn > len(args) {
			fmt.Println("not enough parameters for Context.Do(" + e + ")")
			return nil
		}
		rArgs := make([]reflect.Value, numIn)
		for i := 0; i < numIn; i++ {
			rArgs[i] = reflect.ValueOf(args[i])
		}
		resValue := fn.Call(rArgs)
		if len(resValue) < 1 {
			resSlice = append(resSlice, []interface{}{})
			continue
		}
		res := make([]interface{}, len(resValue))
		for i, v := range resValue {
			res[i] = v.Interface()
		}
		resSlice = append(resSlice, res)
	}
	return resSlice
}

func (this *Context) End() {
	if this.IsEnd {
		return
	}
	if !this.IsSend {
		this.Send()
	}
	this.IsEnd = true
	this.Do(CONTEXT_END)
}

/**
 * 获取或者设置cookie
 */
func (this *Context) Cookie(key string, value ...string) string {
	if len(value) < 1 {
		c, e := this.Request.Cookie(key)
		if e != nil {
			return ""
		}
		return c.Value
	}
	if len(value) == 2 {
		t := time.Now()
		expire, _ := strconv.Atoi(value[1])
		t = t.Add(time.Duration(expire) * time.Second)
		cookie := &http.Cookie{
			Name:    key,
			Value:   value[0],
			Path:    "/",
			MaxAge:  expire,
			Expires: t,
		}
		http.SetCookie(this.Response, cookie)
		return ""
	}
	return ""
}

// GetHeader returns header string by given key.
func (this *Context) GetHeader(key string) string {
	return this.Request.Header.Get(key)
}

// Redirect does redirection response to url string and status int optional.
func (this *Context) Redirect(url string, status ...int) {
	this.Header["Location"] = url
	if len(status) > 0 {
		this.Status = status[0]
		return
	}
	this.Status = 302
}

// ContentType sets content-type string.
func (this *Context) ContentType(contentType string) {
	this.Header["Content-Type"] = contentType
}

// Json set json response with data and proper header.
func (this *Context) Json(data interface{}) {
	bytes, e := json.MarshalIndent(data, "", "    ")
	if e != nil {
		fmt.Println("JSON MARSHAL ERROR: ", e)
	}
	this.ContentType("application/json;charset=UTF-8")
	this.Body = bytes
}

// Send does response sending.
// If response is sent, do not sent again.
func (this *Context) Send() {
	if this.IsSend {
		return
	}
	for name, value := range this.Header {
		this.Response.Header().Set(name, value)
	}
	this.Response.WriteHeader(this.Status)
	this.Response.Write(this.Body)
	this.IsSend = true
	this.Do(CONTEXT_SEND)
}

// Throw throws http status error and error message.
// It call event named as status.
// The context will be end.
func (this *Context) Throw(status int, message ...interface{}) {
	e := strconv.Itoa(status)
	this.Status = status
	this.Do(e, message...)
	this.End()
}

// Tpl returns string of rendering template with data.
// If error, panic.
func (this *Context) Tpl(tpl string, data map[string]interface{}) string {
	b, e := this.app.view.Render(tpl+".tpl", data)
	if e != nil {
		fmt.Println("TPL TEMPLATE ERROR: ", e)
	}
	return string(b)
}

// Render does template and layout rendering with data.
// The result bytes are assigned to context.Body.
// If error, panic.
func (this *Context) Render(tpl string, data map[string]interface{}) {
	b, e := this.app.view.Render(tpl+".tpl", data)
	if e != nil {
		fmt.Println("RENDER ERROR: ", e)
	}
	this.Body = b
	this.Do(CONTEXT_RENDERED)
}

// Func adds template function to view.
// It will affect global *View instance.
func (this *Context) Func(name string, fn interface{}) {
	this.app.view.FuncMap[name] = fn
}

// App returns *App instance in this context.
func (this *Context) App() *App {
	return this.app
}

// Download sends file download response by file path.
func (this *Context) Download(file string) {
	f, e := os.Stat(file)
	if e != nil {
		this.Status = 404
		return
	}
	if f.IsDir() {
		this.Status = 403
		return
	}
	output := this.Response.Header()
	output.Set("Content-Type", "application/octet-stream")
	output.Set("Content-Disposition", "attachment; filename="+path.Base(file))
	output.Set("Content-Transfer-Encoding", "binary")
	output.Set("Expires", "0")
	output.Set("Cache-Control", "must-revalidate")
	output.Set("Pragma", "public")
	http.ServeFile(this.Response, this.Request, file)
	this.IsSend = true
}
