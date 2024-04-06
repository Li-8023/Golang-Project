package models

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"gopkg.in/fatih/set.v0"
	"gorm.io/gorm"
)

// 消息
type Message struct {
	gorm.Model
	FromId   int64  //发送者
	TargetId int64  //消息接收者
	Type     int    //发送类型 （群聊，私聊）
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

	isValidToken := true

	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return isValidToken
		},
	}).Upgrade(writer, request, nil)

	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
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

	sendMsg(userId, []byte("欢迎进入聊天室"))

}

func sendProc(node *Node) {
	for {
		select {
		case data := <-node.DataQueue:
			err := node.Conn.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				fmt.Println("WriteMessage error:", err)
				return
			}
		}
	}

}

func recvProc(node *Node) {
	for {

		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println("ReadMessage error:", err)
			return
		}
		broadMsg(data)
		fmt.Println("[ws] <<<<< ", data)
	}
}

var udpsendChan chan []byte = make(chan []byte, 1024)

func broadMsg(data []byte) {
	udpsendChan <- data
}

func init() {
	go udpSendProc()
	go udpRecProc()
}

// 完成udp数据发送协程
func udpSendProc() {

	con, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 3000,
	})
	// con, err := net.DialUDP("udp", nil, &net.UDPAddr{
	// 	// 10.16.0.179
	// 	// 192.168.100.134
	// 	IP: net.IPv4(192, 168, 100, 134),
	// 	//   IP:   net.IPv4(127,0,0,1),
	// 	Port: 3000,
	// })

	defer con.Close()
	if err != nil {
		fmt.Println(err)
	}

	for {
		select {
		case data := <-udpsendChan:
			_, err := con.Write(data)
			if err != nil {
				fmt.Println("WriteMessage error:", err)
				return
			}
		}
	}
}

// 完成udp数据接受
func udpRecProc() {
	con, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 3000,
	})

	if err != nil {
		fmt.Println("Error", err)
		return
	}
	defer con.Close()

	for {
		var buf [512]byte
		n, err := con.Read(buf[0:])

		if err != nil {
			fmt.Println("Error", err)
			return
		}

		dispatch(buf[0:n])
	}
}

// 后端调度逻辑
func dispatch(data []byte) {
	msg := Message{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	switch msg.Type {

	case 1: //私信
		sendMsg(msg.TargetId, data)
		// case 2: //群发
		// 	sendGroupMsg()
		// case 3: //广播
		// 	sendAllMsg()
		// case 4:

	}
}

func sendMsg(userId int64, msg []byte) {
	rwLocker.RLock()
	node, ok := clientMap[userId]
	rwLocker.RUnlock()

	if ok {
		node.DataQueue <- msg
	}
}
