//Here we handle HTTP requests

package service

import (
	"fmt"
	"ginchat/models"
	"ginchat/utils"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	// "golang.org/x/net/websocket"
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
// @param email query string false "email"
// @param phone query string false "phone"
// @Success 200 {string} json{"code", "message"}
// @Router /user/createUser [get]
func CreateUser(c *gin.Context) {
	user := models.UserBasic{
		Name: c.Query("name"),
	}
	password := c.Query("password")
	repassword := c.Query("repassword")
	email := c.Query("email")
	phone := c.Query("phone")

	salt := fmt.Sprintf("%06d", rand.Int31())

	findName := models.FindUserByName(user.Name)
	if findName.Name != "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户名已注册"})
		return
	}

	findPhone := models.FindUserByPhone(phone)
	if findPhone.Phone != "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该电话已注册"})
		return
	}

	findEmail := models.FindUserByEmail(email)

	if findEmail.Email != "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该邮箱已注册"})
		return
	}
	// Basic input validation
	if user.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户名不能为空"})
		return
	}
	if password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "密码不能为空"})
		return
	}
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "电话不能为空"})
		return
	}
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "邮箱不能为空"})
		return
	}
	if password != repassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "两次密码不一致",
		})
		return
	}

	user.Password = utils.MakePassword(password, salt)
	user.Salt = salt
	user.Email = email
	user.Phone = phone
	if _, err := govalidator.ValidateStruct(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "校验数据失败", "error": err.Error()})
		return
	}

	result := models.CreateUser(user)
	if result.Error != nil {
		// Log the error for debugging
		fmt.Printf("Error creating user: %v\n", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建用户失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "新增用户成功！",
		"info":    result,
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

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "用户删除成功", "info": result})
}

// UpdateUser
// @Summary 更新用户
// @Tags User
// @param id formData string false "id"
// @param name formData string false "name"
// @param password formData string false "password"
// @param phone formData string false "phone"
// @param email formData string false "email"
// @Success 200 {string} json{"code", "message"}
// @Router /user/updateUser [post]
func UpdateUser(c *gin.Context) {
	var user models.UserBasic

	idStr := c.PostForm("id")
	fmt.Println("Received ID:", idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// Additional check to see if idStr is empty
		if idStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "用户ID参数缺失"})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的用户ID"})
		}
		return
	}

	user.ID = uint(id)
	user.Name = c.PostForm("name")
	user.Password = c.PostForm("password")
	user.Phone = c.PostForm("phone")
	user.Email = c.PostForm("email")

	// Validate user struct here
	if _, err := govalidator.ValidateStruct(user); err != nil {
		// If validation fails, return an error message
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "校验数据失败", "error": err.Error()})
		return
	}

	err = models.UpdateUser(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新用户失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "修改用户成功", "info": err})
}

// Login
// @Summary 登录
// @Tags User
// @param name query string false "name"
// @param password query string false "password"
// @Success 200 {string} json{"code", "message"}
// @Router /user/login [post]
func Login(c *gin.Context) {
	name := c.Query("name")
	password := c.Query("password")

	// Attempt to find the user by name
	user := models.FindUserByName(name)

	if user.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该用户不存在"})
		return
	}

	if !utils.ValidPassword(password, user.Salt, user.Password) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "密码不正确"})
		return
	}

	pwd := utils.MakePassword(password, user.Salt)
	data := models.Login(name, pwd)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "登录成功", "info": data})
}

// 防止跨域站点伪造请求
var upGrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

func SendMsg(c *gin.Context) {
	ws, err := upGrade.Upgrade(c.Writer, c.Request, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer func(ws *websocket.Conn) {
		err = ws.Close()
		if err != nil {
			fmt.Println(err)
			return
		}
	}(ws)

	MsgHandler(ws, c)
}

func MsgHandler(ws *websocket.Conn, c *gin.Context) {
	for {
		msg, err := utils.Subscribe(c, utils.PublishKey)
		if err != nil {
			fmt.Println(err)
			return
		}
		tm := time.Now().Format("2006-01-02T15:04:05")
		m := fmt.Sprintf("[ws][%s]:%s", tm, msg)
		err = ws.WriteMessage(1, []byte(m))

		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
