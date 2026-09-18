package router

import (
	docs "chronchat/docs"
	"chronchat/service"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)


func Router() *gin.Engine {
	
	r := gin.Default()
	// swagger
	docs.SwaggerInfo.BasePath = ""
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	//静态资源
	r.Static("/asset","asset/")
	r.LoadHTMLGlob("views/**/*")

	//首页
	r.GET("/", service.GetIndex)
	r.GET("/index", service.GetIndex)
	r.GET("/toRegister",service.ToRegister)
	r.GET("/toChat",service.ToChat)
	r.GET("/user/chat",service.Chat)
	r.POST("/searchFriends",service.SearchFriends)
	//用户模块
	r.POST("/user/getUserList", service.GetUserList)
	r.POST("/user/createUser", service.CreateUser)
	r.POST("/user/deleteUser", service.DeleteUser)
	r.POST("/user/findUserByNameAndPwd", service.FindUserByNameAndPwd)
	r.POST("/user/updateUser", service.UpdateUser)

	//发送消息
	r.GET("/user/sendMsg",service.SendMsg)
	r.GET("/user/sendUserMsg",service.SendUserMsg)
	return r
}
	