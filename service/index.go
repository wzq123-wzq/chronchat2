package service

import (
	"fmt"
	"html/template"
	"strconv"

	"github.com/gin-gonic/gin"

	"chronchat/models"
)

// GetIndex
// @Summary 首页欢迎接口
// @Tags 首页
// @Success 200 {string} welcom
// @Router /index [get]
func GetIndex(c *gin.Context) {
	ind, err := template.ParseFiles("index.html")
	if err != nil {
		panic(err)
	}
	ind.Execute(c.Writer, "index")
}

// ToLogin
// @Summary 登录页
// @Tags 首页
// @Router /user/login.shtml [get]
func ToLogin(c *gin.Context) {
	//login.html使用了define命名块 需要同时解析其引用的head模板 并按名称执行
	ind, err := template.ParseFiles("views/user/login.html", "views/chat/head.html")
	if err != nil {
		panic(err)
	}
	ind.ExecuteTemplate(c.Writer, "/user/login.shtml", nil)
}

// ToRegister
// @Summary 注册页
// @Tags 首页
// @Router /user/register.shtml [get]
func ToRegister(c *gin.Context) {
	ind, err := template.ParseFiles("views/user/register.html")
	if err != nil {
		panic(err)
	}
	ind.Execute(c.Writer, "register")
}

// ToChat
// @Summary 聊天主页
// @Tags 首页
// @Router /chat/index.shtml [get]
func ToChat(c *gin.Context) {
	ind, err := template.ParseFiles("views/chat/index.html",
		"views/chat/head.html",
		"views/chat/tabmenu.html",
		"views/chat/concat.html",
		"views/chat/group.html",
		"views/chat/profile.html",
		"views/chat/main.html",
		"views/chat/createcom.html",
		"views/chat/userinfo.html",
		"views/chat/foot.html")
	if err != nil {
		panic(err)
	}
	userId, _ := strconv.Atoi(c.Query("userId"))
	if userId == 0 {
		//兼容前端登录后的跳转地址 /chat/index.shtml?id=xxx&token=xxx
		userId, _ = strconv.Atoi(c.Query("id"))
	}
	token := c.Query("token")
	user := models.UserBasic{}
	user.ID = uint(userId)
	user.Identity = token
	fmt.Println("userId:", userId, "token:", token)
	ind.Execute(c.Writer, user)
}

// Chat websocket聊天入口
func Chat(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}
