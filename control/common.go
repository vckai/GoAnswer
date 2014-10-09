package control

import (
	"time"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/websocket"

	"github.com/vckai/GoAnswer/model"
)

const (
	writeWait                     = 10 * time.Second
	pongWait                      = 60 * time.Second
	pingPeriod                    = (pongWait * 9) / 10
	messageSize                   = 512
	roomMaxUser                   = 2
	rommGameViews                 = 3  //每局游戏答错次数
	playGameTime    time.Duration = 20 //每局游戏时间
	rebotSubmitTime time.Duration = 2  //机器人答题时间
)

var (
	roomInc     = 1        //房间ID流水号
	rebotUserId = 10000000 //机器人ID流水号
)

type connection struct {
	userId int
	ws     *websocket.Conn
	send   chan []byte
}

/**
 * api返回
 */
type apiParam struct {
	Action string
	UserId int
	Params map[string]interface{}
	Time   int64
}

type onlineUser struct {
	model.Users
	RoomId int
}

type hub struct {
	register      chan *connection
	unregister    chan *connection
	connections   map[int]*connection
	notloginconns map[int]*connection
	broadcast     chan *simplejson.Json
	onlineUsers   map[int]*onlineUser
	rooms         map[int]*room //room list.
	examids       []int
}

var h = hub{
	register:      make(chan *connection),
	unregister:    make(chan *connection),
	connections:   make(map[int]*connection),
	notloginconns: make(map[int]*connection),
	onlineUsers:   make(map[int]*onlineUser),
	broadcast:     make(chan *simplejson.Json),
	rooms:         make(map[int]*room),
}
