package models

import (
	"chronchat/utils"
	"gorm.io/gorm"
	"fmt"
)

// 人员关系
type Contact struct {
	gorm.Model
	OwnerId  uint //谁的关系信息
	TargetId uint //对应的谁
	Type     int  //对应的类型  1好友 2群组 3
	Desc     string
}

func (table *Contact) TableName() string {
	return "contact"
}

func SearchFriends(userId uint) []UserBasic {
	contacts := make([]Contact, 0)
	objIds := make([]uint64, 0)
	err:= utils.DB.Where("owner_id = ? and type = 1", userId).Find(&contacts).Error
	if err != nil {
		fmt.Println("SearchFriends:", err)
	}
	for _,v:= range contacts {
		fmt.Println("SearchFriends:", v)
		objIds = append(objIds, uint64(v.TargetId))
	}
	users := make([]UserBasic, 0)
	utils.DB.Where("id in ?", objIds).Find(&users)
	return users
}
