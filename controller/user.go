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
// @Description 用户登录请求参数
type LoginRequest struct {
	UserName string `json:"username" binding:"required" example:"admin"`        // 用户名
	Password string `json:"password" binding:"required" example:"password123"`  // 密码
}

// LoginResponseData 登录响应数据
// @Description 登录成功响应数据
type LoginResponseData struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // JWT令牌
}

// Login 处理用户登录请求
// @Summary 用户登录
// @Description 用户登录接口，验证用户名和密码，返回JWT令牌
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body LoginRequest true "登录请求参数"
// @Success 200 {object} response.Response{data=LoginResponseData} "登录成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 401 {object} response.Response "登录失败"
// @Router /api/v1/auth/login [post]
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