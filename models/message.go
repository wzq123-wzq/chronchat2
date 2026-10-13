package models

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	"chronchat/utils"
)

// Message 消息表
// 嵌入 gorm.Model 自带字段: ID(主键), CreatedAt(消息发送时间), UpdatedAt(更新时间), DeletedAt(软删除时间)
// json tag 与前端字段约定保持一致:
//   - 发送者ID 前端为 userId
//   - 图片/语音地址 前端为 url
//   - 语音时长 前端为 amount
type Message struct {
	gorm.Model
	FromId   uint   `json:"userId"` // 发送者用户ID,服务端强制覆盖防止伪造
	TargetId uint   // 接收者ID: 私聊时为对方用户ID,群聊时为群ID
	Type     int    // 消息通道类型: 1=私聊 2=群聊 3=心跳 4=广播(预留)
	Media    int    // 消息媒体类型: 1=文字 2=表情包 3=语音 4=图片
	Content  string // 消息文本内容,文字/表情包类消息存这里
	Pic      string // 图片类消息的缩略图地址
	Url      string `json:"url"` // 媒体文件的完整访问URL(图片/语音)
	Desc     string // 附加描述,如图片说明/语音文字摘要等
	Amount   int    `json:"amount"` // 媒体附属数值: 语音时长(秒)/图片像素大小等
}

func (table *Message) TableName() string {
	return "message"
}

// Node 在线用户节点: 一个登录用户对应一个websocket连接
type Node struct {
	Conn      *websocket.Conn // 该用户对应的WebSocket连接对象
	DataQueue chan []byte     // 待发送消息队列,sendProc协程从此消费数据写给客户端
}

// 映射关系: userId -> 在线节点
var clientMap map[uint]*Node = make(map[uint]*Node, 0)

var rwLocker sync.RWMutex

// Chat websocket聊天入口 前端连接地址: ws://host/chat?userId=xxx&token=xxx
func Chat(writer http.ResponseWriter, request *http.Request) {
	//1.获取参数 校验token等合法性
	query := request.URL.Query()
	id := query.Get("userId")
	id2, _ := strconv.ParseUint(id, 10, 0)
	userId := uint(id2)
	// token := query.Get("token")
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
		DataQueue: make(chan []byte, 50), // 存放发送数据的管道
	}
	//3.加载用户群组关系
	//4.userId跟node绑定 并加锁
	rwLocker.Lock()
	clientMap[userId] = node
	rwLocker.Unlock()

	//更新登录时间
	UpdateUserLoginTime(userId)
	//完成发送的逻辑
	go sendProc(node)
	//完成接收的逻辑
	go recvProc(node, userId)
	sendMsg(userId, []byte("欢迎进入聊天室"))
}

// 完成发送的逻辑: 从消息队列取出数据写给客户端
func sendProc(node *Node) {
	for data := range node.DataQueue {
		fmt.Println("[ws]sendProc>>>>>>msg:", string(data))
		err := node.Conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			fmt.Println("sendProc error:", err)
			return
		}
	}
}

// 完成接收的逻辑: 读取客户端消息并分发 连接断开时清理在线表
func recvProc(node *Node, userId uint) {
	for {
		_, data, err := node.Conn.ReadMessage()
		if err != nil {
			fmt.Println("recvProc error:", err)
			RemoveUserConn(userId, node)
			return
		}
		dispatch(data, userId)
		fmt.Println("[ws]recvProc<<<<<", string(data))
	}
}

// RemoveUserConn 用户下线时移除在线映射并关闭连接
func RemoveUserConn(userId uint, node *Node) {
	rwLocker.Lock()
	if cur, ok := clientMap[userId]; ok && cur == node {
		delete(clientMap, userId)
	}
	rwLocker.Unlock()
	err := node.Conn.Close()
	if err != nil {
		fmt.Println("close conn error:", err)
	}
	fmt.Println("user offline, userId:", userId)
}

// 后端调度逻辑: 解析消息并按类型分发
func dispatch(data []byte, userId uint) {
	msg := Message{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	//以websocket连接中的userId为准 防止伪造发送者
	msg.FromId = userId
	switch msg.Type {
	case 1: //私聊
		//消息落库MySQL
		utils.DB.Create(&msg)
		//消息写入Redis 供历史消息拉取
		SortRedisMsg(msg)
		//转发给接收者
		data, err = json.Marshal(msg)
		if err != nil {
			fmt.Println(err)
			return
		}
		sendMsg(msg.TargetId, data)
	case 2: //群聊
		//消息落库MySQL
		utils.DB.Create(&msg)
		//转发给群内所有成员(含发送者 前端群聊消息依赖服务端回显)
		SendGroupMsg(msg)
	case 3: //心跳 更新心跳时间
		now := time.Now()
		utils.DB.Model(&UserBasic{}).Where("id = ?", userId).Update("heartbeat_time", &now)
	case 4: //广播 预留
		return
	}
}

// SendGroupMsg 群聊消息转发 查询群成员逐个投递
func SendGroupMsg(msg Message) {
	contacts := make([]Contact, 0)
	err := utils.DB.Where("target_id = ? AND type = 2", msg.TargetId).Find(&contacts).Error
	if err != nil {
		fmt.Println("SendGroupMsg error:", err)
		return
	}
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, v := range contacts {
		sendMsg(v.OwnerId, data)
	}
}

// SortRedisMsg 将私聊消息按收发双方ID升序作为key写入Redis
func SortRedisMsg(msg Message) {
	min := int64(math.Min(float64(msg.FromId), float64(msg.TargetId)))
	max := int64(math.Max(float64(msg.FromId), float64(msg.TargetId)))
	key := fmt.Sprintf("chat_%d_%d", min, max)
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = utils.Red.RPush(context.Background(), key, data).Err()
	if err != nil {
		fmt.Println("SortRedisMsg error:", err)
	}
}

// WriteRedisMsg 将数据库获取的历史消息按收发双方ID升序作为key写入Redis
func WriteRedisMsg(msg Message) {
	min := int64(math.Min(float64(msg.FromId), float64(msg.TargetId)))
	max := int64(math.Max(float64(msg.FromId), float64(msg.TargetId)))
	key := fmt.Sprintf("chat_%d_%d", min, max)
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = utils.Red.LPush(context.Background(), key, data).Err()
	if err != nil {
		fmt.Println("WriteRedisMsg error:", err)
	}
}

// RedisMsg 拉取两个用户之间的历史消息
// isRev为false时 倒序返回最新的(end-start+1)条 与前端首次加载/上拉分页的渲染顺序一致
func RedisMsg(userIdA, userIdB, start, end int64, isRev bool) ([]string, error) {
	min := int64(math.Min(float64(userIdA), float64(userIdB)))
	max := int64(math.Max(float64(userIdA), float64(userIdB)))
	key := fmt.Sprintf("chat_%d_%d", min, max)
	ctx := context.Background()
	if isRev {
		//正序返回第start到end条
		return utils.Red.LRange(ctx, key, start, end).Result()
	}
	//倒序: 从最新一条往前数 取第start到end条
	length := utils.Red.LLen(ctx, key).Val()
	endIdx := length - 1 - start
	if endIdx < 0 {
		//没有更多历史消息了
		//删除Redis中旧的消息
		utils.Red.Del(ctx, fmt.Sprintf("chat_%d_%d", min, max))
		MySQLMsg(userIdA, userIdB)
		length = utils.Red.LLen(ctx, key).Val()
		endIdx = length - 1 - start
		if endIdx < 0 {
			return []string{}, nil
		}
	}
	startIdx := length - 1 - end
	if startIdx < 0 {
		startIdx = 0
	}
	msgs, err := utils.Red.LRange(ctx, key, startIdx, endIdx).Result()
	if err != nil {
		return msgs, err
	}
	//反转 让最新的消息排在最前面
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

// MySQLMsg 从MySQL查询两个用户之间的历史消息 并写入Redis
func MySQLMsg(userIdA, userIdB int64) {
	msgs := make([]Message, 10)
	err := utils.DB.Where("from_id = ? AND target_id = ? OR from_id = ? AND target_id = ?", userIdA, userIdB, userIdB, userIdA).Order("id DESC").Find(&msgs).Error
	if err != nil {
		fmt.Println("MySQLMsg error:", err)
		return
	}
	//反转 让最新的消息排在最前面
	for _, msg := range msgs {
		WriteRedisMsg(msg)
	}
}

func sendMsg(userId uint, msg []byte) {
	fmt.Println("sendMsg>>>>>userId: ", userId)
	rwLocker.RLock()
	node, ok := clientMap[userId]
	rwLocker.RUnlock()
	if ok {
		node.DataQueue <- msg
	}
}
