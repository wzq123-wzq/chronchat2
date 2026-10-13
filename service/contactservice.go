package service

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"chronchat/models"
	"chronchat/utils"
)

// AddFriend
// @Summary 添加好友
// @Tags 关系
// @Param targetName formData string false "好友用户名"
// @Param userId formData string false "当前用户ID"
// @Success 200 {string} json{"code","msg","data"}
// @Router /contact/addfriend [post]
func AddFriend(c *gin.Context) {
	userId, _ := strconv.Atoi(c.PostForm("userId"))
	targetName := c.PostForm("targetName")
	code, msg := models.AddFriend(uint(userId), targetName)
	if code == 0 {
		utils.RespOK(c.Writer, nil, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

// LoadCommunity
// @Summary 加载用户加入的群聊列表
// @Tags 关系
// @Param ownerId formData string false "当前用户ID"
// @Success 200 {string} json{"code","msg","data"}
// @Router /contact/loadcommunity [post]
func LoadCommunity(c *gin.Context) {
	ownerId, _ := strconv.Atoi(c.PostForm("ownerId"))
	groups := models.LoadCommunity(uint(ownerId))
	utils.RespOKList(c.Writer, groups, len(groups))
}

// CreateCommunity
// @Summary 创建群聊
// @Tags 关系
// @Param ownerId formData string false "群主ID"
// @Param name formData string false "群名称"
// @Param icon formData string false "群头像"
// @Param cate formData string false "群类型"
// @Param desc formData string false "群描述"
// @Success 200 {string} json{"code","msg","data"}
// @Router /contact/createCommunity [post]
func CreateCommunity(c *gin.Context) {
	group := models.GroupBasic{}
	ownerId, _ := strconv.Atoi(c.PostForm("ownerId"))
	group.OwnerId = uint(ownerId)
	group.Name = c.PostForm("name")
	group.Icon = c.PostForm("icon")
	cate, _ := strconv.Atoi(c.PostForm("cate"))
	group.Type = cate
	group.Desc = c.PostForm("desc")
	code, msg := models.CreateCommunity(group)
	if code == 0 {
		utils.RespOK(c.Writer, nil, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}

// JoinGroup
// @Summary 加入群聊(支持群号或群名称)
// @Tags 关系
// @Param comId formData string false "群号或群名称"
// @Param userId formData string false "当前用户ID"
// @Success 200 {string} json{"code","msg","data"}
// @Router /contact/joinGroup [post]
func JoinGroup(c *gin.Context) {
	userId, _ := strconv.Atoi(c.PostForm("userId"))
	comId := c.PostForm("comId")
	code, msg := models.JoinGroup(uint(userId), comId)
	if code == 0 {
		utils.RespOK(c.Writer, nil, msg)
	} else {
		utils.RespFail(c.Writer, msg)
	}
}
