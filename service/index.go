package service

import (
	"chronchat/models"
	"html/template"
	"strconv"
	"fmt"
	"github.com/gin-gonic/gin"
)

// GetIndex
// @Summary 首页欢迎接口
// @Tags 首页
// @Success 200 {string} welcom
// @Router /index [get]
func GetIndex(c *gin.Context) {
	ind,err :=template.ParseFiles("index.html")
	if err !=nil{
		panic(err)
	}
	ind.Execute(c.Writer,"index")
	// c.JSON(200, gin.H{
	// 	"message": "welcom",
	// })
}


func ToRegister(c *gin.Context) {
	ind,err :=template.ParseFiles("views/user/register.html")
	if err !=nil{
		panic(err)
	}
	ind.Execute(c.Writer,"register")
	// c.JSON(200, gin.H{
	// 	"message": "welcom",
	// })
}

func ToChat(c *gin.Context) {
	ind,err :=template.ParseFiles("views/chat/index.html",
	"views/chat/head.html",
	"views/chat/tabmenu.html",
	"views/chat/concat.html",
	"views/chat/group.html",
	"views/chat/profile.html",
	"views/chat/main.html",
	"views/chat/createcom.html",
	"views/chat/userinfo.html",
	"views/chat/foot.html",)
	if err !=nil{
		panic(err)
	}
	userId ,_:= strconv.Atoi(c.Query("userId"))
	token := c.Query("token")
	user := models.UserBasic{}
	user.ID = uint(userId)
	user.Identity = token
	fmt.Println("userId:",userId,"token:",token)
	ind.Execute(c.Writer,user)
}			

func Chat(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}