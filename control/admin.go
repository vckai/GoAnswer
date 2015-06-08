package control

import (
	"fmt"

	"vckai.com/GoAnswer/app"
	"vckai.com/GoAnswer/model"

	"vckai.com/GoAnswer/server"
)

func AdminIndex(context *app.Context) {
	context.Render("admin/index", nil)
}

func AdminExam(context *app.Context) {
	if context.Method == "POST" { // post 请求
		question := context.String("question")
		if len(question) == 0 {
			context.Json(map[string]interface{}{"msg": "请输入问题", "suc": -1})
			context.End()
		}
		options := context.Strings("options[]")
		if len(options) != 4 {
			context.Json(map[string]interface{}{"msg": "需要输入4个答题选项", "suc": -2})
			context.End()
		}
		answer := context.String("answer")
		if len(answer) == 0 {
			context.Json(map[string]interface{}{"msg": "请输入正确答案", "suc": -3})
			context.End()
		}
		answerId := -1
		for key, val := range options {
			if len(val) == 0 {
				context.Json(map[string]interface{}{"msg": "不允许空的答题选项", "suc": -4})
				context.End()
			}
			if val == answer {
				answerId = key
			}
		}
		if answerId == -1 {
			context.Json(map[string]interface{}{"msg": "没有在选项中找到正确答案", "suc": -5})
			context.End()
		}
		resolve := context.MustString("resolve", "")
		_, err := model.AddExam(question, options, int8(answerId), resolve)
		if err != nil {
			fmt.Println("isnert error: ", err)
			context.Json(map[string]interface{}{"msg": "添加失败", "suc": -6})
			context.End()
			return
		}

		// 重新加载题目
		server.SetReload(0)
		context.Json(map[string]interface{}{"msg": "", "suc": 1})
		context.End()
	}
	context.Render("admin/exam", nil)
}
