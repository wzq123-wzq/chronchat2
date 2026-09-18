package service

import (
	"chronchat/models"
	"chronchat/utils"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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
// @Summary 新增用户
// @Tags 用户
// @Param name formData string false "用户名"
// @Param password formData string false "密码"
// @Param repassword formData string false "确认密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/createUser [post]
func CreateUser(c *gin.Context) {
	user := models.UserBasic{}
	fmt.Println("新增用户")
	name := c.PostForm("name")
	password := c.PostForm("password")
	repassword := c.PostForm("Identity")
	salt := fmt.Sprintf("%06d", rand.Int31())

	fmt.Println("名字" + name + ",密码: " + password + ",确认密码: " + repassword)
	if name == "" || password == "" {
		c.JSON(200, gin.H{
			"code":    -1, // 0表示成功，-1表示失败
			"message": "用户名或密码不能为空",
			"data":    user,
		})
		return
	}
	data, err := models.FindUserByName(name)
	if data.ID != 0 && err == nil {
		c.JSON(200, gin.H{
			"code":    -1, // 0表示成功，-1表示失败
			"message": "用户名已注册",
			"data":    user,
		})
		return
	}

	if password != repassword {
		c.JSON(200, gin.H{
			"code":    -1,
			"message": "两次输入的密码不一致",
			"data":    user,
		})
		return
	}
	user.Name = name
	user.Password = utils.MakrePassword(password, salt)
	user.Salt = salt
	fmt.Println(user.Password)
	models.CreateUser(user)
	c.JSON(200, gin.H{
		"code":    0, // 0表示成功，-1表示失败
		"message": "新增用户成功",
		"data":    user,
	})
}

// GetUser
// @Summary 查询用户
// @Tags 用户
// @Param name formData string false "用户名"
// @Param password formData string false "密码"
// @Success 200 {string} json{"code","message"}
// @Router /user/FindUserByNameAndPwd [post]
func FindUserByNameAndPwd(c *gin.Context) {
	data := models.UserBasic{}
	name := c.PostForm("name")
	password := c.PostForm("password")
	fmt.Println("名字" + name + ",密码: " + password)
	user, _ := models.FindUserByName(name)
	if user.Name == "" {
		c.JSON(200, gin.H{
			"code":    -1, // 0表示成功，-1表示失败
			"message": "用户不存在",
			"data":    data,
		})
		return
	}
	falg := utils.ValidPassword(password, user.Salt, user.Password)
	if !falg {
		c.JSON(200, gin.H{
			"code":    -1, // 0表示成功，-1表示失败
			"message": "用户密码错误",
			"data":    data,
		})
		return
	}
	pwd := utils.MakrePassword(password, user.Salt)
	data, _ = models.FindUserByNameAndPwd(name, pwd)
	c.JSON(200, gin.H{
		"code":    0, // 0表示成功，-1表示失败
		"message": "用户登录成功",
		"data":    data,
	})
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
		"code":    0, // 0表示成功，-1表示失败
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

	_, err := govalidator.ValidateStruct(user)
	if err != nil {
		fmt.Println(err)
		c.JSON(400, gin.H{
			"code":    -1, // 0表示成功，-1表示失败
			"message": "修改用户失败",
			"data":    user,
		})
	} else {
		models.UpdateUser(user)
		c.JSON(200, gin.H{
			"code":    0, // 0表示成功，-1表示失败
			"message": "修改用户成功",
			"data":    user,
		})
	}
}

// 防止跨域站点伪造请求
var upGrade = websocket.Upgrader{
	CheckOrigin : func(r *http.Request) bool {
		fmt.Println("5")
		return true
	},
}

func SendMsg(c *gin.Context) {
	ws,err:=upGrade.Upgrade(c.Writer, c.Request, nil)
	fmt.Println("1")
	if err !=nil{
		fmt.Println(err)
		return
	}
	defer func(ws *websocket.Conn){
		err = ws.Close()
		if err !=nil{
			fmt.Println(err)
		}
	}(ws)
	MsgHandler(ws,c)
}

func MsgHandler(ws *websocket.Conn,c *gin.Context){
	fmt.Println("2")
	msg,err :=utils.Subscribe(c,utils.PublishKey)
	if err != nil{
		fmt.Println(err)
	}
	tm:=time.Now().Format("2006-01-02 15:04:05")
	m:=fmt.Sprintf("[ws][%s]:%s",tm,msg)
	err = ws.WriteMessage(1,[]byte(m))
	if err!=nil{
		fmt.Println(err)
	}
}

func SendUserMsg(c *gin.Context) {
	models.Chat(c.Writer,c.Request)
}

func SearchFriends(c *gin.Context) {
	userId, _ := strconv.Atoi(c.PostForm("userId"))
	users := models.SearchFriends(uint(userId))
	// c.JSON(200, gin.H{
	// 	"code":    0,
	// 	"message": "搜索好友成功",
	// 	"data":    users,
	// })
	utils.RespOKList(c.Writer, users,len(users))
	fmt.Println("SearchFriends:", users)
}