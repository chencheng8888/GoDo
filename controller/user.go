package controller

import (
	"fmt"
	"net/http"

	"github.com/chencheng8888/GoDo/auth"
	"github.com/chencheng8888/GoDo/pkg/response"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	auth *auth.Auth
}

func NewUserController(a *auth.Auth) *UserController {
	return &UserController{
		auth: a,
	}
}

// LoginRequest 登录请求结构体
type LoginRequest struct {
	UserName string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponseData 登录响应数据
type LoginResponseData struct {
	Token string `json:"token"`
}

// Login 处理用户登录请求
func (uc *UserController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.InvalidRequestCode, fmt.Sprintf("%s:%s", response.InvalidRequestMsg, err.Error())))
		return
	}

	// 调用auth包的Login方法进行身份验证
	token, err := uc.auth.Login(req.UserName, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Error(response.LoginFailedCode, fmt.Sprintf("%s:%s", response.LoginFailedMsg, err.Error())))
		return
	}

	// 登录成功，返回token
	c.JSON(http.StatusOK, response.Success(LoginResponseData{Token: token}))
}