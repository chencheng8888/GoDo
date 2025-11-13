package controller

import (
	"errors"
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
	UserName string `json:"username" binding:"required" example:"admin"`       // 用户名
	Password string `json:"password" binding:"required" example:"password123"` // 密码
}

// LoginResponseData 登录响应数据
// @Description 登录成功响应数据
type LoginResponseData struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // JWT令牌
}

// RegisterRequest 注册请求结构体
// @Description 用户注册请求参数
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"` // 邮箱
	UserName string `json:"username" binding:"required" example:"user123"`             // 用户名
	Password string `json:"password" binding:"required,min=6" example:"password123"`   // 密码，最少6位
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
// @Failure 401 {object} response.Response "密码错误"
// @Failure 404 {object} response.Response "用户不存在"
// @Failure 500 {object} response.Response "服务器内部错误"
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
		// 根据不同的错误类型返回不同的响应
		if errors.Is(err, auth.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, response.Error(response.UserNotFoundCode, response.UserNotFoundMsg))
			return
		}
		if errors.Is(err, auth.ErrPasswordIncorrect) {
			c.JSON(http.StatusUnauthorized, response.Error(response.PasswordIncorrectCode, response.PasswordIncorrectMsg))
			return
		}
		// 其他错误（如token生成失败、数据库错误等）
		c.JSON(http.StatusInternalServerError, response.Error(response.LoginFailedCode, fmt.Sprintf("%s:%s", response.LoginFailedMsg, err.Error())))
		return
	}

	// 登录成功，返回token
	c.JSON(http.StatusOK, response.Success(LoginResponseData{Token: token}))
}

// Register 处理用户注册请求
// @Summary 用户注册
// @Description 用户注册接口，创建新用户账户
// @Tags 用户认证
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "注册请求参数"
// @Success 200 {object} response.Response "注册成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Failure 409 {object} response.Response "用户名或邮箱已存在"
// @Failure 500 {object} response.Response "服务器内部错误"
// @Router /api/v1/auth/register [post]
func (uc *UserController) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.InvalidRequestCode, fmt.Sprintf("%s:%s", response.InvalidRequestMsg, err.Error())))
		return
	}

	// 调用auth包的Register方法进行用户注册
	err := uc.auth.Register(req.Email, req.UserName, req.Password)
	if err != nil {
		// 根据不同的错误类型返回不同的响应
		if errors.Is(err, auth.ErrEmailExists) {
			c.JSON(http.StatusConflict, response.Error(response.EmailExistsCode, response.EmailExistsMsg))
			return
		}
		if errors.Is(err, auth.ErrUserNameExists) {
			c.JSON(http.StatusConflict, response.Error(response.UserNameExistsCode, response.UserNameExistsMsg))
			return
		}
		if errors.Is(err, auth.ErrPasswordHash) {
			c.JSON(http.StatusInternalServerError, response.Error(response.InternalErrorCode, "密码加密失败"))
			return
		}
		// 其他错误（如数据库错误等）
		c.JSON(http.StatusInternalServerError, response.Error(response.InternalErrorCode, fmt.Sprintf("注册失败:%s", err.Error())))
		return
	}

	// 注册成功
	c.JSON(http.StatusOK, response.Success("注册成功"))
}
