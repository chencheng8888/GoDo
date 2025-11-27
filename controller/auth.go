package controller

import (
	"fmt"
	"github.com/chencheng8888/GoDo/auth"
	"github.com/chencheng8888/GoDo/pkg/response"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthController struct {
	auth *auth.AuthService
}

func NewAuthController(auth *auth.AuthService) *AuthController {
	return &AuthController{
		auth: auth,
	}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponseData struct {
	Token string `json:"token"`
}

// Login 登录
// @Summary 登录
// @Description 登录
// @Tags 鉴权
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body LoginRequest true "登录参数"
// @Success 200 {object} response.Response{data=LoginResponseData} "success"
// @Failure 400 {object} response.Response "invalid request"
// @Failure 401 {object} response.Response "login failed"
// @Failure 500 {object} response.Response "sign token failed"
// @Router /api/v1/auth/login [post]
func (a *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(response.InvalidRequestCode, fmt.Sprintf("%s:%s", response.InvalidRequestMsg, err.Error())))
		return
	}
	err := a.auth.Authenticate(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Error(response.LoginFailedCode, fmt.Sprintf("%s:%s", response.LoginFailedMsg, err.Error())))
		return
	}
	token, err := a.auth.SignJwtToken(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(response.SignTokenFailedCode, response.SignTokenMsg))
		return
	}

	c.JSON(http.StatusOK, response.Success(LoginResponseData{
		Token: token,
	}))
}
