package control

import (
	"errors"
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"labix.org/v2/mgo/bson"
	"github.com/vckai/GoAnswer/model"
)

/**
 * 房间用户信息
 */
type roomUser struct {
	UserId int
	Status int  //用户准备状态
	IsAct  bool //当前答题用户
	Win    int  //用户胜利局数
	Lose   int  //用户失败局数
	Count  int  //用户总局数
	Views  int  //用户剩余答题次数

	IsReBot bool //是否是机器人
}

/**
 * 房间信息
 */
type room struct {
	Id       int        //房间ID
	UserId   int        //房主ID
	Name     string     //房间名称
	Users    []roomUser //房间用户列表
	Status   int        //房间游戏状态
	Rtype    int        //房间类型
	Time     int64      //房间创建时间
	Count    int        //当前游戏总局数
	ExamList []int      //complete exam list
	Answer   chan int   //submit answer.
	exam     model.Exam //当前题库
	Ready    chan bool
	Over     chan bool //离开游戏
	Rebot    bool      //机器人房间
}

func (this *room) run() {
	for {
		select {
		case ready := <-this.Ready:
			if ready {
				go this.playGame(1)
			}
		}
	}
}

/**
 * is ready
 */
func (this *room) isReady() bool {
	fmt.Println(this.Users)
	if this.Status == 1 {
		fmt.Println("状态：游戏中")
		return false
	}
	if len(this.Users) != roomMaxUser {
		fmt.Println("人数不足", roomMaxUser, "人")
		return false
	}
	for _, user := range this.Users {
		if user.Status == 0 {
			fmt.Println("用户", user.UserId, "还未准备")
			return false
		}
	}
	return true
}

/**
 * 开始游戏
 */
func (this *room) playGame(start int) {
	isOk := this.Status == 1
	if !this.isReady() && !isOk {
		return
	}
	this.Status = 1 //game status.
	exam, err := this.getExam()
	this.exam = exam
	this.ExamList = append(this.ExamList, exam.ExamId)
	isRebot := false //是否机器人
	if start == 1 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		uid := r.Intn(len(this.Users)) //随机抽选用户
		this.Users[uid].IsAct = true
		isRebot = this.Users[uid].IsReBot //是否机器人
	} else {
		isAct := false
		for key, user := range this.Users {
			if user.IsAct == false && isAct == false { //当前答题用户
				this.Users[key].IsAct = true
				isAct = true
				isRebot = user.IsReBot //当前答题用户是机器人
			} else if user.IsAct == true {
				this.Users[key].IsAct = false
			}
		}
	}

	for _, user := range this.Users {
		if user.IsReBot { //机器人状态不需要发送
			continue
		}
		h.sendToClient("PlayGame", user.UserId, map[string]interface{}{
			"Exam":     exam,
			"Err":      err,
			"Users":    this.Users,
			"GameTime": playGameTime,
		})
	}
	GameOutTime := playGameTime
	if isRebot { //机器人自动提交答案时间
		GameOutTime = rebotSubmitTime
	}
	timer := time.NewTimer(GameOutTime * time.Second)
	for { //wait submit answer
		select {
		case answerId := <-this.Answer:
			fmt.Println("提交答案", answerId)
			this.submit(answerId)
			runtime.Goexit()
			break
		case <-this.Over: //游戏结束
			fmt.Println("游戏结束")
			runtime.Goexit()
			break
		case <-timer.C: //timeout use systemp auto submit answer.
			fmt.Println("提交答案超时，系统自动提交")
			anwser := -1
			if isRebot { //机器人自动提交答案
				anwser = this.getRandAnswer()
			}
			this.submit(anwser)
			runtime.Goexit()
			break
		}
	}
}

/**
 * 游戏结束
 */
func (this *room) endGame() {
	defer func() {
		this.clearGame()
	}()

	for _, user := range this.Users {
		if user.IsReBot { //机器人状态不需要发送
			continue
		}
		h.sendToClient("EndGame", user.UserId, map[string]interface{}{
			"Users": this.Users,
		})
	}
}

/**
 * 中途退出房间
 */
func (this *room) GameOver(userId int) {
	fmt.Println("退出房间", userId)
	var delkey int
	for key, user := range this.Users {
		if user.UserId == userId {
			this.Users[key].Views = 0
			delkey = key
		} else {
			if user.IsReBot { //机器人状态则直接删除该房间
				h.delRoom(this.Id, user.UserId)
				continue
			}

			if this.Status == 1 { // 游戏中的状态
				fmt.Println("游戏中")
				h.sendToClient("GameOver", user.UserId, map[string]interface{}{
					"OverUser":     userId,
					"OverUserName": h.onlineUsers[userId].Users.UserName,
				})
				this.Over <- true
				this.endGame() //清空游戏状态
			} else {
				h.sendToClient("OutRoom", user.UserId, map[string]interface{}{
					"OverUser":     userId,
					"OverUserName": h.onlineUsers[userId].Users.UserName,
				})
			}
		}
	}
	if delkey >= 0 {
		fmt.Println("将用户", userId, "从房间", this.Id, "中移除")
		this.Users = append(this.Users[:delkey], this.Users[delkey+1:]...)
		if len(this.Users) == 0 {
			this.ExamList = make([]int, 0)
		}
	}
}

/**
 * submit answer
 */
func (this *room) submit(answer int) {
	if this.Status != 1 {
		fmt.Println("不在游戏中，请勿随便提交答案")
	}
	isOk := false
	if this.exam.ExamAnwser == answer {
		isOk = true
	}
	isEnd := false
	userId := 0
	for key, user := range this.Users {
		if user.IsAct == true { //current active user.
			userId = user.UserId
			if isOk == true {
				this.Users[key].Win++
			} else {
				this.Users[key].Lose++
				this.Users[key].Views--         //剩余次数
				if this.Users[key].Views == 0 { //是否结束游戏
					isEnd = true
				}
			}
			this.Users[key].Count++
		}
	}

	this.Count++
	for _, user := range this.Users {
		if user.IsReBot { //机器人状态不需要发送
			continue
		}
		h.sendToClient("GameResult", user.UserId, map[string]interface{}{
			"Answer":     this.exam.ExamAnwser,
			"IsOk":       isOk,
			"UserId":     userId,
			"UserAnswer": answer,
		})
	}

	time.Sleep(3 * time.Second) //延时3秒

	if isEnd == true { //结束游戏
		this.endGame()
	} else { //重新开始游戏
		go this.playGame(0)
	}
}

/**
 * 获取随机答案
 */
func (this *room) getRandAnswer() int {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return r.Intn(len(this.exam.ExamOption))
}

/**
 * 获取当前答题的题目
 */
func (this *room) getExam() (model.Exam, error) {
	examId := this.getRandExamId()
	if examId == 0 { //已经完成所有题目
		return model.Exam{}, errors.New("Not Get Exam.")
	}
	exam, err := model.GetExam(examId)
	if err != nil {
		return model.Exam{}, err
	}
	return exam, nil
}

/**
 * 随机获取题目ID
 */
func (this *room) getRandExamId() int {
	isNotList := false
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	num := len(h.examids)
	for {
		isNotList = false
		examId := h.examids[r.Intn(num)]
		for _, inExamId := range this.ExamList {
			if examId == inExamId {
				isNotList = true
			}
		}
		if isNotList == false {
			return examId
		}
	}
	// for _, examId := range h.examids {
	// 	isNotList = false
	// 	for _, inExamId := range this.ExamList {
	// 		if examId == inExamId {
	// 			isNotList = true
	// 		}
	// 	}
	// 	if isNotList == false {
	// 		return examId
	// 	}
	// }
	return 0
}

/**
 * 创建机器人
 */
func (this *room) createRebot() {
	//r := rand.New(rand.NewSource(time.Now().UnixNano()))
	//userId, _ := r.Int(99999999, 999999999)
	userId := rebotUserId
	u := &onlineUser{
		model.Users{
			Id_:         bson.NewObjectId(),
			UserId:      userId,
			UserName:    fmt.Sprintf("机器人%d", userId),
			UserPwd:     "",
			UserCoin:    0,
			UserWin:     0,
			UserLose:    0,
			UserRegTime: time.Now(),
		},
		this.Id,
	}

	h.onlineUsers[userId] = u //添加到在线用户信息
	rebotU := roomUser{}
	rebotU.UserId = userId
	rebotU.Status = 1
	rebotU.Win = 0
	rebotU.Lose = 0
	rebotU.Count = 0
	rebotU.IsReBot = true
	rebotU.Views = rommGameViews
	this.Users = append(this.Users, rebotU)

	this.Rebot = true //设置为机器人房间
	rebotUserId++
}

/**
 * 游戏结束, 清空状态
 */
func (this *room) clearGame() {
	for key, _ := range this.Users {
		if !this.Users[key].IsReBot { //机器人不清除准备状态
			this.Users[key].Status = 0
		}
		this.Users[key].Count = 0
		this.Users[key].IsAct = false
		this.Users[key].Lose = 0
		this.Users[key].Views = rommGameViews
		this.Users[key].Win = 0
	}
	this.Count = 0
	this.exam = model.Exam{}
	this.Status = 0
}
