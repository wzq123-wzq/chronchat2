package router

import (
	docs "chronchat/docs"
	"chronchat/service"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Router() *gin.Engine {
	// 初始化路由
	r := gin.Default()
	// swagger
	docs.SwaggerInfo.BasePath = ""
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	//静态资源
	r.Static("/asset", "asset/")
	r.LoadHTMLGlob("views/**/*")

	//页面路由(与前端跳转地址保持一致)
	r.GET("/", service.GetIndex)
	r.GET("/index", service.GetIndex)
	r.GET("/user/login.shtml", service.ToLogin)
	r.GET("/user/register.shtml", service.ToRegister)
	r.GET("/toRegister", service.ToRegister)
	r.GET("/toChat", service.ToChat)
	r.GET("/chat/index.shtml", service.ToChat)

	//用户模块
	r.POST("/user/login", service.Login)
	r.POST("/user/getUserList", service.GetUserList)
	r.POST("/user/createUser", service.CreateUser)
	r.POST("/user/deleteUser", service.DeleteUser)
	r.POST("/user/findUserByNameAndPwd", service.FindUserByNameAndPwd)
	r.POST("/user/updateUser", service.UpdateUser)
	r.POST("/user/find", service.UserFind)
	r.POST("/user/redisMsg", service.RedisMsg)

	//关系模块(好友/群组)
	r.POST("/searchFriends", service.SearchFriends)
	r.POST("/contact/addfriend", service.AddFriend)
	r.POST("/contact/loadcommunity", service.LoadCommunity)
	r.POST("/contact/createCommunity", service.CreateCommunity)
	r.POST("/contact/joinGroup", service.JoinGroup)

	//聊天模块
	r.GET("/chat", service.Chat) //前端websocket连接地址: ws://host/chat
	r.GET("/user/chat", service.Chat)
	r.GET("/user/sendMsg", service.SendMsg)
	r.GET("/user/sendUserMsg", service.SendUserMsg)

	//附件上传
	r.POST("/attach/upload", service.Upload) //前端上传地址
	r.POST("/user/upload", service.Upload)

	return r
}
