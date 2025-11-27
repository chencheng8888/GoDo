package auth

import (
	"errors"
	"fmt"
	"github.com/chencheng8888/GoDo/config"
	"github.com/chencheng8888/GoDo/dao"
	"github.com/chencheng8888/GoDo/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/wire"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
)

var ProviderSet = wire.NewSet(NewAuthService)

const ContextUsernameKey = "username"

type AuthService struct {
	userDao         *dao.UserDao
	tokenExpiration time.Duration
	jwtSecret       string
}

func NewAuthService(userDao *dao.UserDao, cf *config.JwtConfig) *AuthService {
	if len(cf.Secret) == 0 {
		panic("jwt secret cannot be empty")
	}
	if cf.TokenExpiration < 0 {
		panic("jwt token expiration cannot be negative")
	}

	return &AuthService{
		userDao:         userDao,
		jwtSecret:       cf.Secret,
		tokenExpiration: time.Duration(cf.TokenExpiration) * time.Second,
	}
}

func (a *AuthService) Authenticate(username, password string) error {
	pwd, err := a.userDao.GetUser(username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("user %s not found", username)
	} else if err != nil {
		return err
	}

	if pwd != password {
		return fmt.Errorf("user %s password error", username)
	}
	return nil
}

type Claims struct {
	Username             string `json:"username"`
	jwt.RegisteredClaims        // 嵌入标准的 JWT 注册声明
}

func (a *AuthService) SignJwtToken(userName string) (string, error) {
	// 1. 设置 JWT 的过期时间
	expirationTime := time.Now().Add(a.tokenExpiration)

	// 2. 创建 Claims（声明）
	claims := &Claims{
		Username: userName,
		RegisteredClaims: jwt.RegisteredClaims{
			// 设置过期时间 (Expiration Time)
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			// 设置签发者 (Issuer)
			Issuer: "GoDo-auth-service",
			// 设置签发时间 (Issued At)
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	// 3. 使用 Claims 和签名方法 (例如 HS256) 创建 Token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 4. 使用定义的密钥 (JwtSecret) 签名 Token，生成最终的字符串
	tokenString, err := token.SignedString([]byte(a.jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token for user %s: %w", userName, err)
	}

	return tokenString, nil
}

// ParseJwtToken 解析并验证 JWT 字符串，成功则返回 Claims，失败则返回错误
func (a *AuthService) ParseJwtToken(tokenString string) (*Claims, error) {
	// 准备一个用于接收解析后的 Claims 实例
	claims := &Claims{}

	// 解析 Token
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法是否是 HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// 返回密钥用于验证签名
		return []byte(a.jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token parsing failed: %w", err)
	}

	// 检查 Token 是否有效（包括签名和过期时间等）
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// 类型断言检查，确保 Claims 是我们定义的类型
	if c, ok := token.Claims.(*Claims); ok {
		return c, nil
	}

	return nil, errors.New("invalid token claims")
}

// AuthMiddleware 创建一个用于验证 JWT 的 Gin 中间件
func AuthMiddleware(authService *AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从请求头获取 Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, response.Error(response.AuthorizationHeaderRequiredCode, response.AuthorizationHeaderRequiredMsg))
			c.Abort()
			return
		}

		// 2. 检查格式是否为 "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.JSON(http.StatusUnauthorized, response.Error(response.BearerRequiredCode, response.BearerRequiredMsg))
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 3. 调用 AuthService 验证 Token
		claims, err := authService.ParseJwtToken(tokenString)
		if err != nil {
			// 解析失败（签名错误、过期等）
			c.JSON(http.StatusUnauthorized, response.Error(response.InvalidTokenCode, response.InvalidTokenMsg))
			c.Abort()
			return
		}

		// 4. 将解析出的 Username 存储到 Context 中
		c.Set(ContextUsernameKey, claims.Username)

		// 5. 继续处理请求
		c.Next()
	}
}

// GetUsernameFromContext 是一个帮助函数，用于从 Context 中获取已认证的用户名
func GetUsernameFromContext(c *gin.Context) (string, bool) {
	username, ok := c.Get(ContextUsernameKey)
	if !ok {
		return "", false
	}
	return username.(string), true
}
