package models

import (
	"encoding/json"
	"ginchat/utils"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"context"
	"time"
	
	"github.com/gorilla/websocket"
	"gopkg.in/fatih/set.v0"
	"gorm.io/gorm"
	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
)

// 消息
type Message struct {
	gorm.Model
	UserId   int64  //发送者
	TargetId int64  //消息接收者
	Type     int    //发送类型 （1私聊，2群聊，3广播）
	Media    int    //消息类型 （1文字，2表情包，3音频，4图片）
	Content  string //消息内容
	CreateTime uint64 //创建时间
	ReadTime uint64 
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
	Addr string
	FirstTime uint64
	HeartbeatTime uint64
	LoginTime uint64
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

	currentTime := uint64(time.Now().Unix())
	//获取connection
	node := &Node{
		Conn:      conn,
		Addr : conn.RemoteAddr().String(),
		HeartbeatTime: currentTime,
		LoginTime: currentTime,
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

	sendMsg(userId, []byte("您好，欢迎进入聊天室"))

}

func sendProc(node *Node) {
	for {
		
		select {
		case data := <-node.DataQueue:
			fmt.Println("sendProc >>> msg: ", string(data))
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

		broadMsg(data) //to-do 将消息广播到局域网
		fmt.Println("[ws] recvProc <<<<< ", string(data))

		// if msg.Type == 3 {
		// 	currentTime := uint64(time.Now().Unix())
		// 	node.Heartbeat(currentTime)
		// } else {
		// 	dispatch(data)
		// 	broadMsg(data) //todo 将消息广播到局域网
		// 	fmt.Println("[ws] recvProc <<<<< ", string(data))
		// }
	}
}

var udpsendChan chan []byte = make(chan []byte, 1024)

func broadMsg(data []byte) {
	udpsendChan <- data
}

func init() {
	go udpSendProc()
	go udpRecProc()
	fmt.Println("Init go routines")
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
		fmt.Println("udpRecProc data: ", string(buf[0:n]))
		dispatch(buf[0:n])
	}
}

// 后端调度逻辑
func dispatch(data []byte) {
	msg := Message{}
	// msg.CreateTime = uint64(time.Now().Unix())
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	switch msg.Type {

	case 1: //私信
		fmt.Println("dispatch msg: ", string(data))
		sendMsg(msg.TargetId, data)
		case 2: //群发
			sendGroupMsg(msg.TargetId, data) 
		// case 3: //广播
		// 	sendAllMsg()
		// case 4:

	}
}

func sendGroupMsg(targetId int64, msg []byte) {
	fmt.Println("开始群发消息")
	userIds := SearchUserByGroupId(uint(targetId))
	for i := 0; i < len(userIds); i++ {
		//排除给自己的
		if targetId != int64(userIds[i]) {
			sendMsg(int64(userIds[i]), msg)
		}
	}
}

func sendMsg(userId int64, msg []byte) {
	fmt.Println("sendMsg msg: ", string(msg))

	rwLocker.RLock()
	node, ok := clientMap[userId]
	rwLocker.RUnlock()
	jsonMsg := Message{}
	json.Unmarshal(msg, &jsonMsg)
	ctx := context.Background()
	targetIdStr := strconv.Itoa(int(userId))
	userIdStr := strconv.Itoa(int(jsonMsg.UserId))
	jsonMsg.CreateTime =uint64( time.Now().Unix())
	r, err := utils.Red.Get(ctx, "online_"+userIdStr).Result()

	if r != "" {
		if ok {
			fmt.Println("sendMsg >>> userID: ", userId, "  msg:", string(msg))
			node.DataQueue <- msg
		}
	}
	var key string
	if userId > jsonMsg.UserId {
		key = "msg_" + userIdStr + "_" + targetIdStr
	} else {
		key = "msg_" + targetIdStr + "_" + userIdStr
	}

	
	res, err := utils.Red.ZRevRange(ctx, key, 0, -1).Result()
	if err != nil {
		fmt.Println(err)
	}
	score := float64(cap(res)) + 1
	ress, e := utils.Red.ZAdd(ctx, key, &redis.Z{score, msg}).Result() //jsonMsg
	//res, e := utils.Red.Do(ctx, "zadd", key, 1, jsonMsg).Result() //备用 后续拓展 记录完整msg
	if e != nil {
		fmt.Println(e)
	}
	fmt.Println(ress)
}

// 需要重写此方法才能完整的msg转byte[]
func (msg Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(msg)
}



// 获取缓存里面的消息
func RedisMsg(userIdA int64, userIdB int64, start int64, end int64, isRev bool) []string {
	rwLocker.RLock()
	//node, ok := clientMap[userIdA]
	rwLocker.RUnlock()
	//jsonMsg := Message{}
	//json.Unmarshal(msg, &jsonMsg)
	ctx := context.Background()
	userIdStr := strconv.Itoa(int(userIdA))
	targetIdStr := strconv.Itoa(int(userIdB))
	var key string
	if userIdA > userIdB {
		key = "msg_" + targetIdStr + "_" + userIdStr
	} else {
		key = "msg_" + userIdStr + "_" + targetIdStr
	}
	//key = "msg_" + userIdStr + "_" + targetIdStr
	//rels, err := utils.Red.ZRevRange(ctx, key, 0, 10).Result()  //根据score倒叙

	var rels []string
	var err error
	if isRev {
		rels, err = utils.Red.ZRange(ctx, key, start, end).Result()
	} else {
		rels, err = utils.Red.ZRevRange(ctx, key, start, end).Result()
	}
	if err != nil {
		fmt.Println(err) //没有找到
	}
	// 发送推送消息
	/**
	// 后台通过websoket 推送消息
	for _, val := range rels {
		fmt.Println("sendMsg >>> userID: ", userIdA, "  msg:", val)
		node.DataQueue <- []byte(val)
	}**/
	return rels
}


// 更新用户心跳
func (node *Node) Heartbeat(currentTime uint64) {
	node.HeartbeatTime = currentTime
	return
}

// 清理超时连接
func CleanConnection(param interface{}) (result bool) {
	result = true
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("cleanConnection err", r)
		}
	}()
	//fmt.Println("定时任务,清理超时连接 ", param)
	//node.IsHeartbeatTimeOut()
	currentTime := uint64(time.Now().Unix())
	for i := range clientMap {
		node := clientMap[i]
		if node.IsHeartbeatTimeOut(currentTime) {
			fmt.Println("心跳超时..... 关闭连接：", node)
			node.Conn.Close()
		}
	}
	return result
}

// 用户心跳是否超时
func (node *Node) IsHeartbeatTimeOut(currentTime uint64) (timeout bool) {
	if node.HeartbeatTime+viper.GetUint64("timeout.HeartbeatMaxTime") <= currentTime {
		fmt.Println("心跳超时。。。自动下线", node)
		timeout = true
	}
	return
}


func JoinGroup(userId uint, comId string) (int, string) {
	if userId == 0 || comId == "" {
        return -1, "无效的用户ID或群ID" 
    }


	contact := Contact{}
	contact.OwnerId = userId
	contact.Type = 2
	community := Community{}

	result := utils.DB.Where("id=? or name=?", comId, comId).Find(&community)
	if result.Error != nil {
        fmt.Printf("数据库查询错误: %v\n", result.Error) 
        return -1, "数据库查询失败" 
    }

	if community.Name == "" {
		return -1, "没有找到群"
	}
	utils.DB.Where("owner_id=? and target_id=? and type =2 ", userId, comId).Find(&contact)
	
	if contact.TargetId != 0 {
		return -1, "已加过此群"
	} else {
		contact.TargetId = community.ID
		utils.DB.Create(&contact)
		return 0, "加群成功"
	}
}



