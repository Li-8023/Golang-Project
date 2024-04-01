package models

import (
	"gorm.io/gorm"
)

// 人员关系
type Contact struct {
	gorm.Model
	OwnerId  uint //主角
	TargetId uint //对应聊天的是谁
	Type     int  //好友关系还是什么关系
	Desc     string
}

func (table *Contact) TableName() string {
	return "contact"
}
