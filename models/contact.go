package models

import (
	"ginchat/utils"
	"fmt"
	"gorm.io/gorm"
)

// 人员关系
type Contact struct {
	gorm.Model
	OwnerId  uint //主角
	TargetId uint //对应聊天的是谁
	Type     int  //好友关系还是什么关系  1好友 2群
	Desc     string
}

func (table *Contact) TableName() string {
	return "contact"
}

func SearchFriend(userId uint) []UserBasic{
	contacts := make([] Contact, 0)
	objIds := make([]uint64, 0)
	result := utils.DB.Where("owner_id = ? and type=1", userId).Find(&contacts)
	
	if result.Error != nil {
        fmt.Println("Error finding contacts:", result.Error)
        return nil
    }

	 for _, v := range contacts {
        fmt.Println("Find contact>>>>", v)
        objIds = append(objIds, uint64(v.TargetId))
    }

	users := make([]UserBasic, 0)
	utils.DB.Where("id in ?", objIds).Find(&users)
	return users

}