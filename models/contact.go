package models

import (
	"fmt"
	"ginchat/utils"

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

//双向添加好友
func AddFriend(userId uint, targetId uint) int{

	fmt.Println("user_id: ", userId, "target_id: ", targetId)

	if userId == 0 || targetId == 0 {
		return -1 // Invalid user or target ID
	}

	if userId == targetId {
        return -2 // Cannot add self as a friend
    }
	// Find the user by ID
    user := FindUserById(userId)
    if user.Salt == "" {
        return -3 // User does not exist
    }

	// Find the target by ID
    target := FindUserById(targetId)
    if target.Salt == "" {
        return -4 // Target does not exist
    }

    var existingContact Contact

    result := utils.DB.Where("owner_id = ? AND target_id = ? and type=1", userId, targetId).First(&existingContact)
    if result.Error == nil && existingContact.ID != 0 {
        return -5 // Friendship already exists
    }

	
	tx := utils.DB.Begin()

	//事务一旦开始，不论期间什么异常，最终都会rollback
	defer func ()  {
		if r := recover(); r != nil{
			tx.Rollback()
		}
	}()
	
    contact := Contact{
        OwnerId:  userId,
        TargetId: targetId,
        Type:     1, 
    }
	createResult := utils.DB.Create(&contact)

	contact2 := Contact{
        OwnerId:  targetId,
        TargetId: userId,
        Type:     1, 
    }
	create2Result := utils.DB.Create(&contact2)

    if createResult.Error != nil {
		tx.Rollback()
        return -7 // Error occurred during database insert operation
    }

	if create2Result.Error != nil {
		tx.Rollback()
		return -7
	}

	tx.Commit()

    return 0 // Successfully added friend
} 