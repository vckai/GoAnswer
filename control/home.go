package control

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/vckai/GoAnswer/app"
	"github.com/vckai/GoAnswer/libs"
	"github.com/vckai/GoAnswer/model"
)

const (
	EXPTIME = 3600 * 24
)

// 首页
func Index(context *app.Context) {
	UserId := context.Cookie("UserId")
	if UserId == "" {
		context.Redirect("/login")
	}
	context.Render("index", map[string]interface{}{
		"Host": context.Host,
	})
}

// 用户登录
func Login(context *app.Context) {
	UserId := context.Cookie("UserId")
	if UserId != "" {
		context.Redirect("/")
	}
	Msg := ""
	if context.Method == "POST" {
		user := context.String("user")
		if len(user) == 0 {
			Msg = "请输入用户名"
		}
		pwd := context.String("pwd")
		if len(Msg) == 0 && len(pwd) == 0 {
			Msg = "请输入用户密码"
		}
		if len(Msg) == 0 { //登录操作
			user, err := model.GetUser(user)
			if err == nil { //用户存在
				if user.UserPwd == libs.GenPwd(pwd) { //登录成功
					context.Cookie("UserId", strconv.Itoa(user.UserId), strconv.Itoa(EXPTIME))
					context.Redirect("/")
					return
				}
			}
			Msg = "用户或者密码错误"
		}
	}
	context.Render("login", map[string]interface{}{
		"Msg": Msg,
	})
}

// 用户退出
func Logout(context *app.Context) {
	context.Cookie("UserId", "", "-3600")
	context.Redirect("/login")
}

// 注册用户
func Register(context *app.Context) {
	Msg := ""

	_ = model.GetStatus()

	if context.Method == "POST" {
		user := context.String("user")
		if len(user) == 0 {
			Msg = "请输入用户名"
		}
		if m, _ := regexp.MatchString("[a-zA-Z0-9]{4,20}", user); !m && len(Msg) == 0 {
			Msg = "用户名只能是4-20位字母或者数字"
		}
		pwd := context.String("pwd")
		if len(pwd) == 0 && len(Msg) == 0 {
			Msg = "请输入密码"
		}
		pwd2 := context.String("pwd2")
		if len(Msg) == 0 && len(pwd2) == 0 {
			Msg = "请输入确认密码"
		}
		if len(Msg) == 0 && pwd != pwd2 {
			Msg = "两次输入密码不一致"
		}
		if m, _ := regexp.MatchString("(.*){6,20}", pwd); !m && len(Msg) == 0 {
			Msg = "密码必须为6-20位之间"
		}
		if len(Msg) == 0 { //注册用户
			_, err := model.GetUser(user)
			if err != nil { //没有被注册, 则写入注册信息
				userId, err := model.AddUser(user, libs.GenPwd(pwd), 0, 0, 0)
				if err == nil {
					_ = context.Cookie("UserId", strconv.Itoa(userId), strconv.Itoa(EXPTIME))
					context.Redirect("/")
				} else {
					fmt.Println(err)
					Msg = "注册失败"
				}
			} else {
				Msg = "该用户名已被注册"
			}
		}
	}

	context.Render("register", map[string]interface{}{
		"Msg": Msg,
	})
}
