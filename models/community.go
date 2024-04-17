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

	// Find the user by ID
    user := FindUserById(community.OwnerId)
    if user.Salt == "" {
        return -1, "用户不存在" 
    }

	if err := utils.DB.Create(&community).Error; err != nil {
		fmt.Println(err)
		return -1, "建群失败"
	}
	return 0, "群创建成功"
}

func LoadCommunity(ownerId uint) ([]*Community, string, int){
	if ownerId == 0 {
        return nil, "无效的用户ID", -1 // "Invalid user ID"
    }

	data := []*Community{}

	result := utils.DB.Where("owner_id = ?", ownerId).Find(&data)

	if result.Error != nil {
        fmt.Printf("数据库查询错误: %v\n", result.Error) 
        return nil, "加载失败", -1 // "Loading failed"
    }

	if len(data) == 0 {
        return nil, "没有找到群组", -1 
    }

	for _, v := range data {
		fmt.Println(v)
	}
	return data, "加载成功", 0
}