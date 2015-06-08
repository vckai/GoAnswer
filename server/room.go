package server

import (
	"fmt"
	"time"

	"github.com/vckai/GoAnswer/model"

	"labix.org/v2/mgo/bson"
)

// 房间相关信息
type room struct {
	Id     uint32 //房间ID
	UserId int    //房主ID
	Name   string //房间名称
	Time   int64  //房间创建时间

	//
	Rebot bool //机器人房间

	Game *Game
}

// 创建房间
func NewRoom(userId int) (*room, error) {
	rid := newRoomId()
	r := &room{
		Id:     rid,
		UserId: userId,
		Name:   fmt.Sprintf("room_%d", rid),
		Time:   time.Now().Unix(),
	}
	r.Game = NewGame(r)

	return r, nil
}

// 用户准备
func (r *room) userReady(userId int) error {
	return r.Game.userReady(userId)
}

// 添加一个玩家到房间中
func (r *room) addPlayer(userId int, arg ...bool) error {
	isRebot := false
	// 是否为机器人
	if len(arg) == 1 {
		isRebot = arg[0]
	}
	err := r.Game.addPlayer(userId, isRebot)

	return err
}

// 创建机器人
func (r *room) createRebot() error {
	rebotId := newRebotId()

	// 创建一个虚拟的在线用户
	u := &onlineUser{
		model.Users{
			Id_:         bson.NewObjectId(),
			UserId:      rebotId,
			UserName:    fmt.Sprintf("机器人%d", rebotId),
			UserPwd:     "",
			UserCoin:    0,
			UserWin:     0,
			UserLose:    0,
			UserRegTime: time.Now(),
		},
		r.Id,
	}

	h.addOnlineUser(u)

	if err := r.addPlayer(rebotId, true); err != nil {
		fmt.Println("添加机器人进入房间失败：", rebotId, err)
		return err
	}

	r.Rebot = true

	return nil
}

func (r *room) Close() {
	r.Game.Close()
}
