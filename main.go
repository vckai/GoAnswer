package main

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/vckai/GoAnswer/app"
	"github.com/vckai/GoAnswer/control"
	"github.com/vckai/GoAnswer/model"
	"github.com/vckai/GoAnswer/server"
)

func main() {

	a := app.NewApp()

	a.Get("/", control.Index)
	a.Route("POST,GET", "/reg/", control.Register)
	a.Route("POST,GET", "/login/", control.Login)
	a.Get("/logout/", control.Logout)
	a.Get("/ws/", control.Ws)

	a.Get("/admin/", control.AdminIndex)
	a.Route("POST,GET", "/addExam/", control.AdminExam)

	a.Static(func(context *app.Context) { //静态文件处理
		static := "public"

		url := strings.TrimPrefix(context.Url, "/")
		if url == "favicon.ico" {
			url = path.Join(static, url)
		}
		if !strings.HasPrefix(url, static) {
			return
		}

		f, e := os.Stat(url)
		if e == nil {
			if f.IsDir() {
				context.Status = 403
				context.End()
				return
			}
		}

		http.ServeFile(context.Response, context.Request, url)
		context.IsEnd = true
	})
	model.NewModel(a.Config().MustValue("murl", "127.0.0.1")) //连接到mongodb

	server.InitServer()

	a.Run()
}
