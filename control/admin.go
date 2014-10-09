package control

import (
	"fmt"

	"github.com/vckai/GoAnswer/libs"
	"github.com/vckai/GoAnswer/model"
)

func AdminIndex(context *libs.Context) {
	context.Render("admin/index", nil)
}

func AdminExam(context *libs.Context) {
	if context.Method == "POST" { // post 请求
		question := context.String("question")
		if len(question) == 0 {
			context.Json(map[string]interface{}{"msg": "Please Enter Question.", "suc": -1})
			context.End()
		}
		options := context.Strings("options[]")
		if len(options) != 4 {
			context.Json(map[string]interface{}{"msg": "Need Enter 4 Option.", "suc": -2})
			context.End()
		}
		answer := context.String("answer")
		if len(answer) == 0 {
			context.Json(map[string]interface{}{"msg": "Please Enter Answer.", "suc": -3})
			context.End()
		}
		answerId := -1
		for key, val := range options {
			if len(val) == 0 {
				context.Json(map[string]interface{}{"msg": "Options hava empty.", "suc": -4})
				context.End()
			}
			if val == answer {
				answerId = key
			}
		}
		if answerId == -1 {
			context.Json(map[string]interface{}{"msg": "Answer Enter Error, Please Enter Options the Answer.", "suc": -5})
			context.End()
		}
		resolve := context.MustString("resolve", "")
		_, err := model.AddExam(question, options, answerId, resolve)
		if err == nil {
			context.Json(map[string]interface{}{"msg": "", "suc": 1})
			context.End()
		} else {
			fmt.Println("isnert error: ", err)
			context.Json(map[string]interface{}{"msg": "Insert Fail", "suc": -6})
			context.End()
		}

	}
	context.Render("admin/exam", nil)
}
