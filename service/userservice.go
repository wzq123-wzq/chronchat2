package service

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"chronchat/models"
	"chronchat/utils"
)

// GetUserList
// @Summary 查询所有用户
// @Tags 用户
// @Success 200 {string} json{"code","message"}
// @Router /user/getUserList [get]
func GetUserList(c *gin.Context) {
	data := make([]*models.UserBasic, 10)
	data, _ = models.GetUserList()
	c.JSON(200, gin.H{
		"code":    0,
		"message": "查询所有用户成功",
		"data":    data,
	})
}

// CreateUser
// @Summary 新增用户(注册)
// @Tags 用户
// @Param name formData string false "用户名"
// @Param password formData string false "密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/createUser [post]
func CreateUser(c *gin.Context) {
	user := models.UserBasic{}
	fmt.Println("新增用户")
	name := c.PostForm("name")
	password := c.PostForm("password")
	repassword := c.PostForm("repassword")
	salt := fmt.Sprintf("%06d", rand.Int31())

	fmt.Println("名字" + name + ",密码: " + password)
	if name == "" || password == "" {
		c.JSON(200, gin.H{
			"code":    -1, // 0表示成功，-1表示失败
			"message": "用户名或密码不能为空",
			"data":    user,
		})
		return
	}
	//两次密码一致性校验（前端用 repassword 字段名，与 UserBasic.Identity 彻底解耦）
	if password != repassword {
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "两次输入的密码不一致",
			"data":    user,
		})
		return
	}
	data, err := models.FindUserByName(name)
	if data.ID != 0 && err == nil {
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "用户名已注册",
			"data":    user,
		})
		return
	}
	user.Name = name
	user.Password = utils.MakrePassword(password, salt)
	user.Salt = salt
	//生成唯一标识(登录token) —— 服务端自行生成，不依赖前端传入
	user.Identity = utils.Md5Encode(fmt.Sprintf("%s%d", name, time.Now().Unix()))
	fmt.Println(user.Password)
	if err := models.CreateUser(user).Error; err != nil {
		fmt.Println("CreateUser error:", err)
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "注册失败",
			"data":    user,
		})
		return
	}
	c.JSON(200, gin.H{
		"code":    0,
		"message": "新增用户成功",
		"data":    user,
	})
}

// Login
// @Summary 用户登录
// @Tags 用户
// @Param mobile formData string false "账号(用户名或手机号)"
// @Param passwd formData string false "密码"
// @Success 200 {string} json{"code","msg","data"}
// @Router /user/login [post]
func Login(c *gin.Context) {
	mobile := c.PostForm("mobile")
	passwd := c.PostForm("passwd")
	if mobile == "" || passwd == "" {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "账号或密码不能为空",
		})
		return
	}
	//用户名或手机号登录
	user, err := models.FindUserByName(mobile)
	if err != nil || user.ID == 0 {
		user, err = models.FindUserByPhone(mobile)
	}
	if err != nil || user.ID == 0 {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "用户不存在",
		})
		return
	}
	flag := utils.ValidPassword(passwd, user.Salt, user.Password)
	if !flag {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "用户密码错误",
		})
		return
	}
	pwd := utils.MakrePassword(passwd, user.Salt)
	data, err := models.FindUserByNameAndPwd(user.Name, pwd)
	if err != nil {
		c.JSON(200, gin.H{
			"code": -1,
			"msg":  "登录失败",
		})
		return
	}
	//更新登录时间
	models.UpdateUserLoginTime(data.ID)
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "登录成功",
		"data": gin.H{
			"id":     data.ID,
			"token":  data.Identity,
			"name":   data.Name,
			"avatar": data.Avatar,
		},
	})
}

// GetUser
// @Summary 查询用户
// @Tags 用户
// @Param name formData string false "用户名"
// @Param password formData string false "密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/findUserByNameAndPwd [post]
func FindUserByNameAndPwd(c *gin.Context) {
	data := models.UserBasic{}
	name := c.PostForm("name")
	password := c.PostForm("password")
	fmt.Println("名字" + name + ",密码: " + password)
	user, _ := models.FindUserByName(name)
	if user.Name == "" {
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "用户不存在",
			"data":    data,
		})
		return
	}
	falg := utils.ValidPassword(password, user.Salt, user.Password)
	if !falg {
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "用户密码错误",
			"data":    data,
		})
		return
	}
	pwd := utils.MakrePassword(password, user.Salt)
	data, _ = models.FindUserByNameAndPwd(name, pwd)
	c.JSON(200, gin.H{
		"code":    0,
		"message": "用户登录成功",
		"data":    data,
	})
}

// UserFind
// @Summary 查询单个用户信息
// @Tags 用户
// @Param userId formData string false "用户ID"
// @Success 200 {string} json{"code","msg","data"}
// @Router /user/find [post]
func UserFind(c *gin.Context) {
	userId, _ := strconv.Atoi(c.PostForm("userId"))
	data, err := models.FindUserById(uint(userId))
	if err != nil || data.ID == 0 {
		utils.RespFail(c.Writer, "用户不存在")
		return
	}
	utils.RespOK(c.Writer, data, "查询用户成功")
}

// DeleteUser
// @Summary 删除用户
// @Tags 用户
// @Param id query string false "id"
// @Success 200 {string} json{"code","message"}
// @Router /user/deleteUser [get]
func DeleteUser(c *gin.Context) {
	user := models.UserBasic{}
	ID, _ := strconv.Atoi(c.Query("id"))
	user.ID = uint(ID)
	models.DeleteUser(user)
	c.JSON(200, gin.H{
		"code":    0,
		"message": "删除用户成功",
		"data":    user,
	})
}

// UpdateUser
// @Summary 更新用户
// @Tags 用户
// @Param id formData string false "id"
// @Param name formData string false "用户名"
// @Param password formData string false "密码"
// @Param phone formData string false "手机号"
// @Param email formData string false "邮箱"
// @Param icon formData string false "头像"
// @Success 200 {string} json{"code","message"}
// @Router /user/updateUser [post]
func UpdateUser(c *gin.Context) {
	user := models.UserBasic{}
	ID, _ := strconv.Atoi(c.PostForm("id"))
	user.ID = uint(ID)
	user.Name = c.PostForm("name")
	user.Password = c.PostForm("password")
	user.Phone = c.PostForm("phone")
	user.Email = c.PostForm("email")
	user.Avatar = c.PostForm("icon")

	//非空字段才做格式校验 避免只改昵称/头像时被空手机号卡住
	if user.Phone != "" || user.Email != "" {
		_, err := govalidator.ValidateStruct(user)
		if err != nil {
			fmt.Println(err)
			utils.RespFail(c.Writer, "修改用户失败: 手机号或邮箱格式错误")
			return
		}
	}
	models.UpdateUser(user)
	utils.RespOK(c.Writer, user, "修改用户成功")
}

// RedisMsg
// @Summary 拉取历史消息
// @Tags 消息
// @Param userIdA formData string false "用户A的ID"
// @Param userIdB formData string false "用户B的ID"
// @Param start formData int false "开始位置"
// @Param end formData int false "结束位置"
// @Param isRev formData bool false "是否正序"
// @Success 200 {string} json{"code","msg","data"}
// @Router /user/redisMsg [post]
func RedisMsg(c *gin.Context) {
	userIdA, _ := strconv.Atoi(c.PostForm("userIdA"))
	userIdB, _ := strconv.Atoi(c.PostForm("userIdB"))
	start, _ := strconv.Atoi(c.PostForm("start"))
	end, _ := strconv.Atoi(c.PostForm("end"))
	isRev, _ := strconv.ParseBool(c.PostForm("isRev"))
	msgs, err := models.RedisMsg(int64(userIdA), int64(userIdB), int64(start), int64(end), isRev)
	if err != nil {
		utils.RespFail(c.Writer, "拉取历史消息失败")
		return
	}
	//前端从Total中逐条JSON.parse渲染
	utils.RespOKList(c.Writer, "ok", msgs)
}

// 防止跨域站点伪造请求
var upGrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func SendMsg(c *gin.Context) {
	ws, err := upGrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer func(ws *websocket.Conn) {
		err = ws.Close()
		if err != nil {
			fmt.Println(err)
		}
	}(ws)
	MsgHandler(ws, c)
}

// MsgHandler 消息处理函数
func MsgHandler(ws *websocket.Conn, c *gin.Context) {
	msg, err := utils.Subscribe(c, utils.PublishKey)
	if err != nil {
		fmt.Println(err)
	}
	tm := time.Now().Format("2006-01-02 15:04:05")
	m := fmt.Sprintf("[ws][%s]:%s", tm, msg)
	err = ws.WriteMessage(1, []byte(m))
	if err != nil {
		fmt.Println(err)
	}
}

// SendUserMsg 发送用户消息
func SendUserMsg(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}

func SearchFriends(c *gin.Context) {
	userId, _ := strconv.Atoi(c.PostForm("userId"))
	users := models.SearchFriends(uint(userId))
	utils.RespOKList(c.Writer, users, len(users))
	fmt.Println("SearchFriends:", users)
}
