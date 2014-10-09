package main

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/vckai/GoAnswer/control"
	"github.com/vckai/GoAnswer/libs"
	"github.com/vckai/GoAnswer/model"
)

func main() {

	app := libs.NewApp()

	app.Get("/", control.Index)
	app.Route("POST,GET", "/reg/", control.Register)
	app.Route("POST,GET", "/login/", control.Login)
	app.Get("/logout/", control.Logout)
	app.Get("/ws/", control.Ws)

	app.Get("/admin/", control.AdminIndex)
	app.Route("POST,GET", "/addExam/", control.AdminExam)

	app.Static(func(context *libs.Context) { //静态文件处理
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
	model.NewModel(app.Config().MustValue("murl", "127.0.0.1")) //连接到mongodb

	control.InitWs()

	app.Run()
}
