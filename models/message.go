package models

import (
	"gopkg.in/fatih/set.v0"
	"gorm.io/gorm"
)

// 消息
type Message struct {
	gorm.Model
	FormId   uint   //发送者
	TargetId uint   //消息接收者
	Type     string //发送类型 （群聊，私聊）
	Media    int    //消息类型 （文字，图片，音频）
	Content  string //消息内容
	Pic      string
	Url      string
	Desc     string
	Amount   int //其他数字统计

}

func (table *Message) TableName() string {
	return "message"
}

//发送消息需要
//1. 发送者id， 接收者id， 消息类型，发送类型， 发送的内容

type Node struct {
	Conn      *websocket.Conn
	DataQueue chan []byte
	GroupSets set.Interface
}

// 映射关系
var clientMap map[int64]*Node = make(map[int64]*Node, 0)

// 读写锁
var rwLocker sync.RWMutex

func Chat(writer http.ResponseWriter, request *http.Request) {
	//检验token等合法性
	query := request.URL.Query()
	Id := query.Get("userId")
	userId, _ := strconv.ParseInt(Id, 10, 64)
	// token := query.Get("token")
	targetId := query.Get("targetId")
	context := query.Get("context")
	msgType := query.Get("type")
	isvalid := true //checkToken()
	conn, err := (&websocket.Upgrader{
		//token校验
		CheckOrigin: func(r *http.Request) bool {
			return isvalid
		},
	}).Upgrade(writer, request, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	//获取connection
	node := &Node{
		Conn:      conn,
		DataQueue: make(chan []byte, 50),
		GroupSets: set.New(set.ThreadSafe),
	}

	//用户关系
	//userid 跟node绑定，并且加锁
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()

	//完成发送逻辑
	go sendProc(node)
	//完成接受逻辑
	go recvProc(node)
	
}
