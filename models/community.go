package models

import (
	"fmt"
	"ginchat/utils"

	"gorm.io/gorm"
)


type Community struct{
	gorm.Model
	Name string //群名称
	OwnerId uint
	Img string
	Desc string
}

func (table *Community) TableName() string {
	return "community"
}

//创建群
func CreateCommunity(community Community) (int, string){
	if len(community.Name) == 0 {
		return -1, "群名称不能为空"
	}
	if community.OwnerId == 0{
		return -1, "请先登录"
	}

    var existingCommunity Community
    if err := utils.DB.Where("name = ?", community.Name).First(&existingCommunity).Error; err == nil {
        return -1, "群名称已被使用" 
	}

	// // Find the user by ID
    // user := FindUserById(community.OwnerId)
    // if user.Salt == "" {
    //     return -1, "用户不存在" 
    // }

	if err := utils.DB.Create(&community).Error; err != nil {
		fmt.Println(err)
		return -1, "建群失败"
	}
	return 0, "群创建成功"
}