//Here we handle HTTP requests

package service

import (
	"fmt"
	"ginchat/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// GetUserList
// @Summary 所有用户
// @Tags User
// @Success 200 {string} json{"code", "message"}
// @Router /user/getUserList [get]
func GetUserList(c *gin.Context) {

	// data := make([]*models.UserBasic, 10)
	// data = models.GetUserList()
	// models.GetUserList()
	data := models.GetUserList()

	c.JSON(200, gin.H{
		"message": data,
	})
}

// CreateUser
// @Summary 新增用户
// @Tags User
// @param name query string false "UserName"
// @param password query string false "Password"
// @param repassword query string false "Re-enterPassword"
// @Success 200 {string} json{"code", "message"}
// @Router /user/createUser [get]
func CreateUser(c *gin.Context) {
	user := models.UserBasic{
		Name: c.Query("name"),
	}
	password := c.Query("password")
	repassword := c.Query("repassword")

	// Basic input validation
	if user.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户名不能为空"})
		return
	}
	if password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "密码不能为空"})
		return
	}

	if password != repassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "两次密码不一致",
		})
		return
	}

	user.Password = password

	result := models.CreateUser(user)
    if result.Error != nil {
        // Log the error for debugging
        fmt.Printf("Error creating user: %v\n", result.Error)
        c.JSON(http.StatusInternalServerError, gin.H{
            "code": 500,
            "message": "创建用户失败",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "code": 200,
        "message": "新增用户成功！",
    })
}

// DeleteUser
// @Summary 删除用户
// @Tags User
// @param id query string false "id"
// @Success 200 {string} json{"code", "message"}
// @Router /user/deleteUser [get]
func DeleteUser(c *gin.Context) {
	// Check if the ID query parameter exists
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "需要提供用户ID"})
		return
	}

	// Convert ID from string to uint
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户ID"})
		return
	}

	// Create a UserBasic instance with the specified ID for deletion
	userToDelete := models.UserBasic{Model: gorm.Model{ID: uint(id)}}

	// Attempt to delete the user
	result := models.DeleteUser(userToDelete)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "无法删除用户"})
		}
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "用户未找到"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "用户删除成功"})
}
