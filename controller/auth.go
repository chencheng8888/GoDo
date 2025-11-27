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
		c.JSON(http.StatusInternalServerError, response.Error(response.SignTokenFailed, response.SignTokenMsg))
		return
	}

	c.JSON(http.StatusOK, response.Success(LoginResponseData{
		Token: token,
	}))
}
