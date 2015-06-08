package control

import (
	"fmt"
	"strconv"

	"github.com/vckai/GoAnswer/app"
	"github.com/vckai/GoAnswer/server"

	"github.com/gorilla/websocket"
)

// 建立websocket连接
func Ws(context *app.Context) {

	UserId := context.Cookie("UserId")
	if UserId == "" {
		fmt.Println("该用户没有登录")
		context.Throw(403, "用户尚未登录")
		return
	}
	nUserId, _ := strconv.Atoi(UserId)
	if nUserId < 1 {
		fmt.Println("UserID转换失败")
		context.Throw(403, "用户ID错误")
		return
	}

	if _, err := server.GetServer().GetOnlineUser(nUserId); err != server.ErrNotExistsOnlineUser {
		fmt.Println("该用户已经连接过了")
		context.Throw(403, "该用户已经连接过了")
		return
	}
	ws, err := websocket.Upgrade(context.Response, context.Request, nil, 1024, 1024)
	if errmsg, ok := err.(websocket.HandshakeError); ok {
		fmt.Println("websocket连接失败，---ERROR：", errmsg)
		context.Throw(403, "Websocket Not handler")
		return
	} else if err != nil {
		fmt.Println("Websocket连接失败，---ERROR：", err)
		context.Throw(403, "WEBSOCKET连接失败")
		return
	}

	c, err := server.NewConn(nUserId, ws)
	if err != nil {
		fmt.Println("创建Conn失败", err)
		context.Throw(403, "创建Conn失败")
		return
	}

	go c.WritePump()
	c.ReadPump()
}
