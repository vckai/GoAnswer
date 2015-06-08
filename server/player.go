package server

import (
	"sync"
)

// 房间用户信息
type player struct {
	UserId int
	Status bool  //用户准备状态(true准备中, false尚未准备)
	IsAct  bool  //当前答题用户(true为当前答题的用户)
	Win    int16 //用户胜利局数
	Lose   int16 //用户失败局数
	Count  int16 //用户总局数
	Views  int16 //用户剩余答题次数

	IsRebot bool //是否是机器人
	lock    sync.Mutex
}

// 创建一个玩家
// userId  玩家ID
// isRebot 是否机器人
func NewPlayer(userId int, isRebot bool) (*player, error) {
	return &player{
		UserId:  userId,
		Status:  false,
		Win:     0,
		Lose:    0,
		Count:   0,
		Views:   0,
		IsRebot: isRebot,
	}, nil
}

// 设置用户答题状态
func (p *player) act(act bool) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.IsAct = act
}

// 用户准备状态处理
func (p *player) ready(status bool) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.Status = status
}

// 用户房间总局数递增
func (p *player) count() {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.Count++
}

// 失败局数递增
func (p *player) lose() {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.Lose++
}

// 胜利局数递增
func (p *player) win() {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.Win++
}

// 当局剩余游戏次数
func (p *player) views() {
	p.Views--
}
