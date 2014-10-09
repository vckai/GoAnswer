package control

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/websocket"

	"github.com/vckai/GoAnswer/libs"
	"github.com/vckai/GoAnswer/model"
)

func InitWs() {
	h.examids, _ = model.GetAllExamId()
	//fmt.Println(h.examids)
	go h.run()
}

/**
 * 开始运行
 */
func (this *hub) run() {
	for {
		select {
		case c := <-this.register:
			go this.reg(c)
		case c := <-this.unregister:
			go this.logout(c)
		case param := <-this.broadcast:
			go this.do(param)
		}
	}
}

/**
 * 接口执行
 */
func (this *hub) do(param *simplejson.Json) {
	userId, err := param.Get("UserId").Int()
	if err != nil || userId < 1 {
		fmt.Println("传入参数错误，UserID：", userId, "---ERROR：", err)
		return
	}

	if param.Get("Action").MustString() != "Login" {
		c := this.getClient(userId)
		if c == nil {
			fmt.Println("尚未登录, 无法使用该接口：", param)
			return
		}
	}

	switch param.Get("Action").MustString() {
	case "Login": //登录
		this.login(param)
	case "JoinRoom": //加入房间
		this.joinRoom(param)
	case "Submit": //提交答案
		this.submitAnswer(param)
	case "OutRoom": //退出房间
		this.outRoom(param)
	case "Ready": //用户准备
		this.ready(param)
	}
}

/**
 * 获取socket链接
 */
func (this *hub) getClient(userId int) *connection {
	c, ok := this.connections[userId]
	if !ok {
		fmt.Println("没有获取到该socket")
		return nil
	}
	return c
}

/**
 * 注册socket连接
 */
func (this *hub) reg(c *connection) {
	this.notloginconns[c.userId] = c
}

/**
 * 初始化登录
 */
func (this *hub) login(param *simplejson.Json) {
	userId, _ := param.Get("UserId").Int()

	c, ok := this.notloginconns[userId]
	if !ok {
		fmt.Println("该用户尚未建立连接")
		return
	}
	user, err := model.GetUserById(userId)
	if err != nil {
		fmt.Println("没有查找到该用户", err)
		return
	}

	delete(this.notloginconns, userId)
	this.connections[userId] = c

	u := &onlineUser{user, 0} //在线用户信息, 房间ID

	this.onlineUsers[userId] = u
	fmt.Println("用户ID：", userId, "，姓名：", user.UserName, "，登录成功")
}

/**
 * 退出房间
 */
func (this *hub) outRoom(param *simplejson.Json) {
	userId := param.Get("UserId").MustInt()
	fmt.Println("用户", userId, "退出房间")
	if this.onlineUsers[userId].RoomId > 0 {
		this.getRoom(this.onlineUsers[userId].RoomId).GameOver(userId)
		this.onlineUsers[userId].RoomId = 0

		this.sendToClient("OutRoom", userId, map[string]interface{}{
			"OverUser":     userId,
			"OverUserName": this.onlineUsers[userId].Users.UserName,
		})
	}
}

/**
 * 退出socket登录连接, 各种清理
 */
func (this *hub) logout(c *connection) {
	fmt.Println("关闭一个Socket连接", c)
	if this.onlineUsers[c.userId].RoomId > 0 {
		this.getRoom(this.onlineUsers[c.userId].RoomId).GameOver(c.userId)
	}
	delete(this.connections, c.userId)
	delete(this.onlineUsers, c.userId)
	close(c.send)
}

/**
 * 进入房间
 */
func (this *hub) joinRoom(param *simplejson.Json) {
	userId := param.Get("UserId").MustInt()

	isRebot := param.Get("Params").Get("IsRebot").MustInt()

	if this.onlineUsers[userId].RoomId > 0 { //该用户已经加入房间中了
		fmt.Println("禁止重复进入房间", userId)
		return
	}

	rm := new(room)
	roomId := this.findRoom()

	if roomId == 0 { //没有合适的房间则新建一个
		rm.Id = roomInc
		rm.Name = fmt.Sprintf("Room_%d", rm.Id)
		rm.UserId = userId
		rm.Status = 0
		rm.Time = time.Now().Unix()
		rm.Answer = make(chan int)
		rm.Ready = make(chan bool)
		rm.Over = make(chan bool)

		this.rooms[rm.Id] = rm
		roomId = rm.Id
		fmt.Println("创建房间", roomId)

		roomInc = roomInc + 1 //room id update
		go rm.run()
	} else {
		rm = this.getRoom(roomId)
	}

	roomU := roomUser{}
	roomU.UserId = userId
	roomU.Status = 1
	roomU.Win = 0
	roomU.Lose = 0
	roomU.Count = 0
	roomU.Views = rommGameViews

	rm.Users = append(rm.Users, roomU)
	if isRebot == 1 {
		rm.createRebot()
	}

	var users []*onlineUser
	for _, user := range rm.Users { //返回给客户端的用户信息
		users = append(users, this.onlineUsers[user.UserId])
	}
	this.onlineUsers[userId].RoomId = rm.Id

	rm.Ready <- true //用户准备

	for _, user := range rm.Users {
		if user.IsReBot { //机器人状态不需要发送
			continue
		}
		c := this.getClient(user.UserId)
		c.send <- this.genRes("JoinRoom", userId, map[string]interface{}{
			"Room": map[string]interface{}{
				"Id":     rm.Id,
				"Name":   rm.Name,
				"UserId": rm.UserId,
				"Status": rm.Status,
				"Time":   rm.Time,
				"Users":  rm.Users,
			},
			"Users": users,
		})
	}
}

/**
 * submit answer
 */
func (this *hub) submitAnswer(param *simplejson.Json) {
	userId := param.Get("UserId").MustInt()
	user := this.onlineUsers[userId]
	this.rooms[user.RoomId].Answer <- param.Get("Params").Get("AnswerId").MustInt()
}

/**
 * game user ready
 */
func (this *hub) ready(param *simplejson.Json) {
	userId := param.Get("UserId").MustInt()
	user := this.onlineUsers[userId]
	currKey := 0
	for key, user := range this.rooms[user.RoomId].Users {
		if userId == user.UserId {
			currKey = key
		} else {
			if user.IsReBot { //机器人不需要发送
				continue
			}
			this.sendToClient("Ready", user.UserId, map[string]interface{}{
				"UserId": userId,
			})
		}
	}
	this.rooms[user.RoomId].Users[currKey].Status = 1
	this.rooms[user.RoomId].Ready <- true
}

/**
 * 查找未满人的房间
 */
func (this *hub) findRoom() int {
	for _, val := range this.rooms {
		if val.Status == 0 && len(val.Users) < roomMaxUser && val.Rebot == false {
			return val.Id
		}
	}
	return 0
}

/**
 * 获取房间信息
 */
func (this *hub) getRoom(roomId int) *room {
	for _, val := range this.rooms {
		if val.Id == roomId {
			return val
		}
	}
	return nil
}

/**
 * 房间列表
 */
func (this *hub) getRooms() map[int]*room {
	return this.rooms
}

/**
 * 删除某个房间
 */
func (this *hub) delRoom(roomId, rebUserId int) {
	fmt.Println("删除房间", roomId, "删除机器人信息", roomId)
	delete(this.rooms, roomId)
	delete(this.onlineUsers, rebUserId) //删除机器人用户信息
}

/**
 * 发送消息到客户端
 */
func (this *hub) sendToClient(action string, userId int, data map[string]interface{}) {
	c := this.getClient(userId)
	c.send <- this.genRes(action, userId, data)
}

/**
 * 统一返回的数据结构
 */
func (this *hub) genRes(action string, userId int, params map[string]interface{}) []byte {
	v := &apiParam{Action: action, UserId: userId, Params: params, Time: time.Now().Unix()}
	s, err := json.Marshal(v)
	if err != nil {
		fmt.Println("生成JSON数据错误, ERR: ", err)
		return []byte{}
	}
	return []byte(s)
}

/**
 * 建立websocket连接
 */
func Ws(context *libs.Context) {

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
	if _, ok := h.onlineUsers[nUserId]; ok {
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

	c := &connection{userId: nUserId, ws: ws, send: make(chan []byte, 256)}
	h.register <- c

	go c.writePump()
	c.readPump()
}
