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

type Message struct {
	gorm.Model
	FromId   uint
	TargetId uint
	Type     int    //发送类型 1私聊  2群聊  3广播
	Media    int    //消息类型 1文字 2表情包 3图片 4音频
	Content  string //消息内容
	Pic      string
	Url      string
	Desc     string
	Amount   int //其他数字统计
}

func (table *Message) TableName() string {
	return "message"
}

type Node struct {
	Conn      *websocket.Conn
	DataQueue chan []byte
	GroupSets set.Interface
}

// 映射关系
var clientMap map[uint]*Node = make(map[uint]*Node, 0)

var rwLocker sync.RWMutex

func Chat(writer http.ResponseWriter, request *http.Request) {
	//1。获取参数，校验token等合法性
	//token := query.Get("token")
	query := request.URL.Query()
	id := query.Get("userId")
	id2, _ := strconv.ParseUint(id, 10, 0)
	userId := uint(id2)
	// targetId := query.Get("targetId")
	// context := query.Get("context")
	// msgtype := query.Get("type")
	isvalida := true
	conn, err := (&websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return isvalida
		},
	}).Upgrade(writer, request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	//2.获取conn
	node := &Node{
		Conn:      conn,
		DataQueue: make(chan []byte, 50),//存放发送数据的管道
		GroupSets: set.New(set.ThreadSafe),//存放该用户加入的群组ID
	}
	//3.用户关系
	//4.userid跟node绑定 并加锁
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()

	//完成发送的逻辑
	go sendProc(node)
	//完成接收的逻辑
	go recvProc(node)
	sendMsg(userId, []byte("欢迎进入聊天室"))
}


// 完成发送的逻辑
func sendProc(node *Node) {
	for data := range node.DataQueue {
		fmt.Println("[ws]sendProc>>>>>>msg:",string(data))
		err := node.Conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			fmt.Println("sendProc error:", err)
			return
		}
	}
}

// 完成接收的逻辑
func recvProc(node *Node) {
	for {
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println(err)
			return
		}
		broadMsg(data)
		fmt.Println("[ws]recvProc<<<<<", string(data))
	}
}

var udpsendChan chan []byte = make(chan []byte, 1024)

func broadMsg(data []byte) {
	udpsendChan <- data
}

func init() {
	go udpSendProc()
	go udpRecvProc()
	fmt.Println("init goroutine")
}

// 完成udp数据发送协程
func udpSendProc() {
	con, err := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.IPv4(172, 28, 30, 75),
		Port: 3000,
	})
	if err != nil {
		fmt.Println("udp send proc error:", err)
	}
	defer con.Close()
	for data := range udpsendChan {
		fmt.Println("[udp]udpSendProc data : ", string(data))
		_, err := con.Write(data)
		if err != nil {
			fmt.Println("[udp]udp send proc error:", err)
			return
		}
	}
}

// 完成udp数据接收协程
func udpRecvProc() {
	con, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 3000,
	})
	if err != nil {
		fmt.Println(err)
	}
	defer con.Close()
	for {
		var buf [512]byte
		n, err := con.Read(buf[0:])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("[udp]udpRecvProc data : ", string(buf[0:n]))
		dispatch(buf[0:n])
	}
}

// 后端调度逻辑
func dispatch(data []byte) {
	msg := Message{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	switch msg.Type {
	case 1: //私信
		fmt.Println("dispatch data:", string(data))
		sendMsg(msg.TargetId, data)
	// case 2://群发
	// 	sendGroupMsg()
	// case 3://广播
	// 	sendAllMsg()
	case 4:
		return
	}
}

func sendMsg(userId uint, msg []byte) {
	fmt.Println("sendMsg>>>>>userId: ", string(userId))
	rwLocker.RLock()
	node, ok := clientMap[userId]
	rwLocker.RUnlock()
	if ok {
		node.DataQueue <- msg
	}
}
