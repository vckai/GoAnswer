package server

import (
	"sync"
	"sync/atomic"
	"time"

	"vckai.com/GoAnswer/model"

	"github.com/bitly/go-simplejson"
	"github.com/gorilla/websocket"
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
	roomInc     uint32 = 1        //房间ID流水号
	rebotUserId int64  = 10000000 //机器人ID流水号
)

// 连接管理
type Connection struct {
	userId int
	ws     *websocket.Conn
	send   chan []byte
}

// api返回
type apiParam struct {
	Action string
	UserId int
	Params map[string]interface{}
	Time   int64
}

// 在线用户
type onlineUser struct {
	model.Users
	RoomId uint32
}

type hub struct {
	reload        chan int8
	register      chan *Connection
	unregister    chan *Connection
	connections   map[int]*Connection
	notloginconns map[int]*Connection
	broadcast     chan *simplejson.Json
	onlineUsers   map[int]*onlineUser
	rooms         map[uint32]*room //room list.
	examids       []int

	lock sync.Mutex
}

var h = &hub{
	reload:        make(chan int8),
	register:      make(chan *Connection),
	unregister:    make(chan *Connection),
	connections:   make(map[int]*Connection),
	notloginconns: make(map[int]*Connection),
	onlineUsers:   make(map[int]*onlineUser),
	broadcast:     make(chan *simplejson.Json),
	rooms:         make(map[uint32]*room),
}

// 生成房间ID
func newRoomId() uint32 {
	atomic.AddUint32(&roomInc, 1)

	return roomInc
}

// 生成机器人流水ID
func newRebotId() int {
	atomic.AddInt64(&rebotUserId, 1)

	return int(rebotUserId)
}

// 初始化一个新连接
func NewConn(userId int, ws *websocket.Conn) (*Connection, error) {
	c := &Connection{
		userId: userId,
		ws:     ws,
		send:   make(chan []byte, 256),
	}

	h.register <- c

	return c, nil
}

func SetReload(level int8) {
	h.reload <- level
}
