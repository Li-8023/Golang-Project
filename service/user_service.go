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
// @Router /user/getUserList [post]
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
// @Accept x-www-form-urlencoded
// @Produce json
// @Param name formData string true "UserName"
// @Param password formData string true "Password"
// @Param Identity formData string true "Re-enter Password"
// @Param email formData string true "Email"
// @Param phone formData string true "Phone"
// @Success 200 {object} map[string]interface{} "Returns a message on successful user creation"
// @Failure 400 {object} map[string]interface{} "Returns a code and a message if there is a bad request"
// @Failure 500 {object} map[string]interface{} "Returns a code and a message if there is an internal server error"
// @Router /user/createUser [post]
func CreateUser(c *gin.Context) {
	user := models.UserBasic{}

	user.Name = c.Request.FormValue("name")
	password := c.Request.FormValue("password")
	repassword := c.Request.FormValue("Identity")
	email := c.Request.FormValue("email")
	phone := c.Request.FormValue("phone")

	salt := fmt.Sprintf("%06d", rand.Int31())

	findName := models.FindUserByName(user.Name)
	if findName.Name != "" {
		c.JSON(200, gin.H{
			"code": http.StatusBadRequest,
			"message": "用户名已注册"})
		return
	}

	findPhone := models.FindUserByPhone(phone)
	if findPhone.Phone != "" {
		c.JSON(200, gin.H{
			"code": http.StatusBadRequest,
			"message": "该电话已注册"})
		return
	}

	findEmail := models.FindUserByEmail(email)

	if findEmail.Email != "" {
		c.JSON(200, gin.H{
			"code": http.StatusBadRequest,
			"message": "该邮箱已注册"})
		return
	}
	// Basic input validation
	if user.Name == "" {
		c.JSON(200, gin.H{
			"code": http.StatusBadRequest,
			"message": "用户名不能为空"})
		return
	}
	if password == "" {
		c.JSON(200, gin.H{
			"code": http.StatusBadRequest,
			"message": "密码不能为空"})
		return
	}
	if phone == "" {
		c.JSON(200, gin.H{
			"code": http.StatusBadRequest,
			"message": "电话不能为空"})
		return
	}
	if email == "" {
		c.JSON(200, gin.H{
			"code": http.StatusBadRequest,
			"message": "邮箱不能为空"})
		return
	}
	if password != repassword {
		c.JSON(200, gin.H{
			"code": http.StatusBadRequest,
			"message": "两次密码不一致",
		})
		return
	}

	user.Password = utils.MakePassword(password, salt)
	user.Salt = salt
	user.Email = email
	user.Phone = phone
	if _, err := govalidator.ValidateStruct(user); err != nil {
		c.JSON(200, gin.H{
			"code": http.StatusBadRequest,
			"message": "校验数据失败"})
		return
	}

	result := models.CreateUser(user)
	if result.Error != nil {
		// Log the error for debugging
		fmt.Printf("Error creating user: %v\n", result.Error)
		c.JSON(200, gin.H{
			"code": http.StatusInternalServerError,
			"message": "创建用户失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"message": "新增用户成功！",
	})
}

// DeleteUser
// @Summary 删除用户
// @Tags User
// @param id query string false "id"
// @Success 200 {object} map[string]interface{} "Returns a message on successful user creation"
// @Failure 400 {object} map[string]interface{} "Returns a code and a message if there is a bad request"
// @Failure 404 {object} map[string]interface{} "Returns a code and a message if there is a status not found error"
// @Failure 500 {object} map[string]interface{} "Returns a code and a message if there is an internal server error"
// @Router /user/deleteUser [post]
func DeleteUser(c *gin.Context) {
	// Check if the ID query parameter exists
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(200, gin.H{"code": http.StatusBadRequest, "message": "需要提供用户ID"})
		return
	}

	// Convert ID from string to uint
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(200, gin.H{"code": http.StatusBadRequest, "message": "无效的用户ID"})
		return
	}

	// Create a UserBasic instance with the specified ID for deletion
	userToDelete := models.UserBasic{Model: gorm.Model{ID: uint(id)}}

	// Attempt to delete the user
	result := models.DeleteUser(userToDelete)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(200, gin.H{"code": http.StatusNotFound, "message": "用户不存在"})
		} else {
			c.JSON(200, gin.H{"code": http.StatusInternalServerError, "message": "无法删除用户"})
		}
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(200, gin.H{"code": http.StatusNotFound, "message": "用户未找到"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "用户删除成功"})
}

// UpdateUser
// @Summary 更新用户
// @Tags User
// @param id formData string false "id"
// @param name formData string false "name"
// @param password formData string false "password"
// @param phone formData string false "phone"
// @param email formData string false "email"
// @Success 200 {object} map[string]interface{} "Returns a message on successful user creation"
// @Failure 400 {object} map[string]interface{} "Returns a code and a message if there is a bad request"
// @Failure 404 {object} map[string]interface{} "Returns a code and a message if there is a status not found error"
// @Failure 500 {object} map[string]interface{} "Returns a code and a message if there is an internal server error"
// @Router /user/updateUser [post]
func UpdateUser(c *gin.Context) {
	var user models.UserBasic

	idStr := c.PostForm("id")
	fmt.Println("Received ID:", idStr)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		// Additional check to see if idStr is empty
		if idStr == "" {
			c.JSON(200, gin.H{"code": http.StatusBadRequest, "message": "用户ID参数缺失"})
		} else {
			c.JSON(200, gin.H{"code": http.StatusBadRequest, "message": "无效的用户ID"})
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
		c.JSON(200, gin.H{"code": http.StatusBadRequest, "message": "校验数据失败", "error": err.Error()})
		return
	}

	err = models.UpdateUser(user)
	if err != nil {
		c.JSON(200, gin.H{"code": http.StatusInternalServerError, "message": "更新用户失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "修改用户成功"})
}


// Login
// @Summary 登录
// @Tags User
// @Accept x-www-form-urlencoded
// @Produce json
// @Param name formData string true "UserName"
// @Param password formData string true "Password"
// @Success 200 {object} map[string]interface{} "Returns a message on successful user creation"
// @Failure 400 {object} map[string]interface{} "Returns a code and a message if there is a bad request"
// @Failure 404 {object} map[string]interface{} "Returns a code and a message if there is a status not found error"
// @Failure 500 {object} map[string]interface{} "Returns a code and a message if there is an internal server error"
// @Router /user/login [post]
func Login(c *gin.Context) {

	name := c.Request.FormValue("name")
	password := c.Request.FormValue("password")

	// Attempt to find the user by name
	user := models.FindUserByName(name)

	if user.Name == "" {
		c.JSON(200, gin.H{"code": http.StatusBadRequest, "message": "该用户不存在"})
		return
	}

	if !utils.ValidPassword(password, user.Salt, user.Password) {
		c.JSON(200, gin.H{"code": http.StatusBadRequest, "message": "密码不正确"})
		return
	}

	pwd := utils.MakePassword(password, user.Salt)
	data := models.Login(name, pwd)
	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "登录成功", "data": data})
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

func SendUserMsg(c *gin.Context) {
	models.Chat(c.Writer, c.Request)
}
