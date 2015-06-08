package server

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/vckai/GoAnswer/model"
)

var (
	ErrNotExam = errors.New("已经完成了所有题目")
)

type Game struct {
	rm     *room     //房间
	Users  []*player //房间用户列表
	Status int8      //房间状态(0等待1游戏中2满人)
	Count  int8      //当前游戏总局数

	exam      model.Exam //当前题目
	ExamList  []int      //已完成的题库列表
	Answer    chan int8  //开始游戏chan
	GameStart chan bool  //chan
	Over      chan bool  //离开游戏chan

	lock sync.Mutex
}

// new game
func NewGame(r *room) *Game {
	game := &Game{
		Status:    0,
		Count:     0,
		Answer:    make(chan int8),
		GameStart: make(chan bool),
		Over:      make(chan bool),
		rm:        r,
		ExamList:  make([]int, 0),
	}

	go func(game *Game) {
		for {
			fmt.Println("33333333")
			if ok := <-game.GameStart; ok {
				fmt.Println("zhixingdaozheli")
				game.playGame()
			}
		}
	}(game)

	return game
}

// 添加一个玩家到房间中
func (game *Game) addPlayer(userId int, isRebot bool) error {

	player, err := NewPlayer(userId, isRebot)
	if err != nil {
		fmt.Println("创建游戏用户失败：", err)
		return err
	}
	player.Views = rommGameViews

	game.Users = append(game.Users, player)

	var users []*onlineUser
	for _, user := range game.Users { //返回给客户端的用户信息
		u, err := h.GetOnlineUser(user.UserId)
		if err != nil {
			fmt.Println("获取在线用户失败：", err)
			continue
		}
		users = append(users, u)
	}

	game.send("JoinRoom", map[string]interface{}{
		"Room": map[string]interface{}{
			"Id":     game.rm.Id,
			"Name":   game.rm.Name,
			"UserId": userId,
			"Status": game.Status,
			"Time":   game.rm.Time,
			"Users":  game.Users,
		},
		"Users": users,
	})
	// 用户进入房间首次自动准备
	game.userReady(userId)

	return nil
}

// 用户准备
func (game *Game) userReady(userId int) error {
	isStart := false
	for _, user := range game.Users {
		if userId == user.UserId {
			user.ready(true)
		} else {
			isStart = user.Status
		}
	}

	game.send("Ready", map[string]interface{}{
		"UserId": userId,
	})
	if isStart {
		game.GameStart <- true
	}
	return nil
}

// 判断用户是否准备状态
func (game *Game) checkIsReady() bool {

	if len(game.Users) != roomMaxUser {
		fmt.Println("人数不足", roomMaxUser, "人")
		return false
	}
	for _, user := range game.Users {
		if user.Status == false {
			fmt.Println("用户", user.UserId, "还未准备")
			return false
		}
	}

	return true
}

// 开始游戏
func (game *Game) playGame() {
	if !game.checkIsReady() {
		return
	}
	exam, err := game.getExam()
	if err != nil {
		fmt.Println("没有找到题目")
		// game.send("NotExam", map[string]interface{}{
		// 	"Err": err.Error(),
		// })
		game.endGame()
		return
	}
	game.exam = exam
	// 添加该题进入过滤slice
	game.ExamList = append(game.ExamList, exam.ExamId)

	isRebot := false      //是否机器人
	if game.Status == 0 { // 开始一局游戏
		game.Status = 1 //更改房间游戏状态

		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		uid := r.Intn(len(game.Users)) //随机抽选用户
		game.Users[uid].act(true)      //设置当前用户为答题用户

		isRebot = game.Users[uid].IsRebot //当前答题用户是否为机器人
	} else { // 一局游戏中的继续游戏
		isAct := false
		for _, user := range game.Users {
			if user.IsAct == false && isAct == false { //当前答题用户
				user.act(true)
				isAct = true
				isRebot = user.IsRebot //当前答题用户是否为机器人
			} else if user.IsAct == true {
				user.act(false)
			}
		}
	}

	game.send("PlayGame", map[string]interface{}{
		"Exam":     exam,
		"Users":    game.Users,
		"GameTime": playGameTime,
	})

	GameOutTime := playGameTime
	if isRebot { //机器人自动提交答案时间
		GameOutTime = rebotSubmitTime
	}
	timer := time.NewTimer(GameOutTime * time.Second)
	for { //wait submit answer
		select {
		case answerId := <-game.Answer:
			fmt.Println("提交答案", answerId)
			game.submit(answerId)
			return
		case <-game.Over: //游戏结束
			fmt.Println("游戏结束")
			game.endGame()
			return
		case <-timer.C: //游戏超时未提交答案
			fmt.Println("超时未答题")
			var anwser int8 = -1
			if isRebot { //机器人自动提交答案
				anwser = game.getRandAnswer()
			}
			game.submit(anwser)
			return
		}
	}
}

// 中途退出房间
func (game *Game) GameOver(userId int) {
	fmt.Println("退出房间", userId)
	var delkey int
	for key, user := range game.Users {
		if user.UserId == userId {
			game.Users[key].Views = 0
			delkey = key
		} else {
			if user.IsRebot { //
				continue
			}

			u, err := h.GetOnlineUser(userId)
			if err != nil {
				fmt.Println("获取用户失败：", err)
				continue
			}

			if game.Status == 1 { // 游戏中的状态
				game.Over <- true
				h.sendToClient("GameOver", user.UserId, map[string]interface{}{
					"OverUser":     userId,
					"OverUserName": u.Users.UserName,
				})
			} else {
				h.sendToClient("OutRoom", user.UserId, map[string]interface{}{
					"OverUser":     userId,
					"OverUserName": u.Users.UserName,
				})
			}
		}
	}
	if delkey >= 0 {
		fmt.Println("将用户", userId, "从房间", game.rm.Id, "中移除")
		game.Users = append(game.Users[:delkey], game.Users[delkey+1:]...)
		if len(game.Users) == 0 {
			game.ExamList = make([]int, 0)
		}
	}
}

// 提交答案
func (game *Game) submit(answer int8) {
	if game.Status != 1 {
		fmt.Println("不在游戏中，请勿随便提交答案")
		return
	}
	isOk := false
	// 是否答对
	if game.exam.ExamAnwser == answer {
		isOk = true
	}
	isEnd := false
	userId := 0
	for _, user := range game.Users {
		if user.IsAct == true { //current active user.
			if isOk == true { //胜利
				user.win()
			} else { //失败的逻辑处理
				user.lose()          //失败局数递增
				user.views()         //剩余次数递减
				if user.Views <= 0 { //是否结束游戏
					isEnd = true
				}
			}
		}
		user.count() //用户游戏局数递增
	}

	game.Count++
	game.send("GameResult", map[string]interface{}{
		"Answer":     game.exam.ExamAnwser,
		"IsOk":       isOk,
		"UserId":     userId,
		"UserAnswer": answer,
	})

	time.Sleep(3 * time.Second) //延时3秒, 让客户端等待缓冲

	if isEnd == true { //结束一局游戏
		game.endGame()
	} else { //重新开始游戏
		game.playGame()
	}
}

// 获取随机答案
// 机器人答题时使用
func (game *Game) getRandAnswer() int8 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return int8(r.Intn(len(game.exam.ExamOption)))
}

// 获取当前答题的题目
// 从mongodb中读取
//
// @TODO 考虑全部放到内存中
func (game *Game) getExam() (model.Exam, error) {
	examId := game.getRandExamId()
	if examId == 0 { //已经完成所有题目
		return model.Exam{}, ErrNotExam
	}

	exam, err := model.GetExam(examId)
	if err != nil {
		return model.Exam{}, err
	}
	return exam, nil
}

// 随机获取题目ID
// 过滤掉该房间已经完成过的题目
func (game *Game) getRandExamId() int {
	isNotList := false
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	num := len(h.examids)
	if num == 0 || len(h.examids) == len(game.ExamList) {
		return 0
	}
	for {
		isNotList = false
		examId := h.examids[r.Intn(num)]
		for _, inExamId := range game.ExamList {
			if examId == inExamId {
				isNotList = true
			}
		}
		if isNotList == false {
			return examId
		}
	}
	return 0
}

// 往房间用户的客户端发送消息
// 机器人不会发送
func (game *Game) send(action string, res map[string]interface{}) {
	for _, user := range game.Users {
		if user.IsRebot || user.UserId == 0 { //机器人状态不需要发送
			continue
		}
		h.sendToClient(action, user.UserId, res)
	}
}

// 游戏结束
func (game *Game) endGame() {
	game.clearGame()

	game.send("EndGame", map[string]interface{}{
		"Users": game.Users,
	})
}

// 游戏结束, 清空状态
func (game *Game) clearGame() {
	fmt.Println("清空房间信息")

	// 重置用户游戏状态
	for _, u := range game.Users {
		if !u.IsRebot { //机器人不清除准备状态
			u.Status = false
		}
		u.Count = 0
		u.IsAct = false
		u.Lose = 0
		u.Views = rommGameViews
		u.Win = 0
	}

	// 重置游戏状态
	game.Count = 0
	game.exam = model.Exam{}
	game.Status = 0
}

// 关闭游戏, 清除数据
func (game *Game) Close() {
	close(game.Answer)
	close(game.GameStart)
	close(game.Over)
}
