package models

import "gorm.io/gorm"

// GroupBasic 群聊表
// 嵌入 gorm.Model 自带字段: ID(主键,即群ID), CreatedAt(创建时间), UpdatedAt(更新时间), DeletedAt(软删除时间)
type GroupBasic struct {
	gorm.Model
	Name    string // 群名称
	OwnerId uint   // 群主用户ID,创建群聊时自动成为群主
	Icon    string // 群头像URL地址
	Type    int    // 群类型,预留扩展用(如1普通群 2讨论组 3公众号)
	Desc    string // 群描述/群公告
}

func (table *GroupBasic) TableName() string {
	return "group_basic"
}
