package models

import (
	"fmt"
	"strconv"

	"chronchat/utils"
)

// Contact 人员关系表
// 用于描述 用户-好友 或 用户-群组 的从属关系,双向维护保证好友关系对称
type Contact struct {
	OwnerId  uint `json:"userId"` // 关系所有者的用户ID,即"我"
	TargetId uint            // 关系目标ID,好友时为对方用户ID,群组时为群ID
	Type     int             // 关系类型: 1=好友 2=群组
	Desc     string          // 关系备注,如群内昵称/管理员身份标记等
}

func (table *Contact) TableName() string {
	return "contact"
}

// SearchFriends 查询好友列表
func SearchFriends(userId uint) []UserBasic {
	contacts := make([]Contact, 0)
	objIds := make([]uint64, 0)
	err := utils.DB.Where("owner_id = ? AND type = 1", userId).Find(&contacts).Error
	if err != nil {
		fmt.Println("SearchFriends:", err)
	}
	for _, v := range contacts {
		fmt.Println("SearchFriends:", v)
		objIds = append(objIds, uint64(v.TargetId))
	}
	users := make([]UserBasic, 0)
	utils.DB.Where("id IN ?", objIds).Find(&users)
	return users
}

// AddFriend 添加好友 双向写入关系表 事务保证一致性
func AddFriend(userId uint, targetName string) (int, string) {
	if userId == 0 {
		return -1, "未登录"
	}
	tUser, err := FindUserByName(targetName)
	if err != nil || tUser.ID == 0 {
		return -1, "该用户不存在"
	}
	if tUser.ID == userId {
		return -1, "不能添加自己为好友"
	}
	//是否已添加过
	cnt := make([]Contact, 0, 1)
	utils.DB.Where("owner_id = ? AND target_id = ? AND type = 1", userId, tUser.ID).Find(&cnt)
	if len(cnt) > 0 {
		return -1, "该好友已添加"
	}
	//双方都添加进各自的关系表
	tx := utils.DB.Begin()
	contact := Contact{
		OwnerId:  userId,
		TargetId: tUser.ID,
		Type:     1,
	}
	contact1 := Contact{
		OwnerId:  tUser.ID,
		TargetId: userId,
		Type:     1,
	}
	if err := tx.Create(&contact).Error; err != nil {
		tx.Rollback()
		return -1, "添加好友失败"
	}
	if err := tx.Create(&contact1).Error; err != nil {
		tx.Rollback()
		return -1, "添加好友失败"
	}
	tx.Commit()
	return 0, "添加成功"
}

// CreateCommunity 创建群聊 群主自动加入
func CreateCommunity(group GroupBasic) (int, string) {
	if group.OwnerId == 0 {
		return -1, "未登录"
	}
	if len(group.Name) == 0 {
		return -1, "群名称不能为空"
	}
	err := utils.DB.Create(&group).Error
	if err != nil {
		return -1, "建群失败"
	}
	//群主加入群聊
	owner := Contact{
		OwnerId:  group.OwnerId,
		TargetId: group.ID,
		Type:     2,
		Desc:     "群主",
	}
	err = utils.DB.Create(&owner).Error
	if err != nil {
		return -1, "建群失败"
	}
	return 0, "建群成功"
}

// LoadCommunity 加载用户加入的所有群聊
func LoadCommunity(ownerId uint) []GroupBasic {
	contacts := make([]Contact, 0)
	objIds := make([]uint64, 0)
	utils.DB.Where("owner_id = ? AND type = 2", ownerId).Find(&contacts)
	for _, v := range contacts {
		objIds = append(objIds, uint64(v.TargetId))
	}
	groups := make([]GroupBasic, 0)
	utils.DB.Where("id IN ?", objIds).Find(&groups)
	return groups
}

// JoinGroup 通过群号或群名称加入群聊
func JoinGroup(userId uint, comId string) (int, string) {
	if userId == 0 {
		return -1, "未登录"
	}
	group := GroupBasic{}
	//优先按群号查找
	id, err := strconv.ParseUint(comId, 10, 64)
	if err == nil {
		utils.DB.First(&group, id)
	}
	//群号无效时按群名称查找
	if group.ID == 0 {
		utils.DB.Where("name = ?", comId).First(&group)
	}
	if group.ID == 0 {
		return -1, "群聊不存在"
	}
	//是否已加入
	cnt := make([]Contact, 0, 1)
	utils.DB.Where("owner_id = ? AND target_id = ? AND type = 2", userId, group.ID).Find(&cnt)
	if len(cnt) > 0 {
		return -1, "已加入该群聊"
	}
	contact := Contact{
		OwnerId:  userId,
		TargetId: group.ID,
		Type:     2,
	}
	err = utils.DB.Create(&contact).Error
	if err != nil {
		return -1, "加群失败"
	}
	return 0, "加群成功"
}
