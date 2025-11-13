package auth

import (
    "errors"
    "fmt"
    "net/http"
    "strings"
    "time"

    "github.com/chencheng8888/GoDo/config"
    "github.com/chencheng8888/GoDo/dao"
    "github.com/chencheng8888/GoDo/model"
    "github.com/chencheng8888/GoDo/pkg/response"
    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v4"
    "github.com/google/wire"
    "golang.org/x/crypto/bcrypt"
)

var (
    ProviderSet = wire.NewSet(NewAuth)
)

type Auth struct {
    userDao   *dao.UserDao
    jwtSecret []byte
}

// NewAuth 创建Auth实例
func NewAuth(userDao *dao.UserDao, jwtConfig *config.JWTConfig) *Auth {
    return &Auth{
        userDao:   userDao,
        jwtSecret: []byte(jwtConfig.Secret),
    }
}

var (
    ErrUserNotFound      = errors.New("user not found")
    ErrPasswordIncorrect = errors.New("password incorrect")
    ErrTokenGeneration   = errors.New("failed to generate token")
    ErrUserExists        = errors.New("user already exists")
    ErrEmailExists       = errors.New("email already exists")
    ErrUserNameExists    = errors.New("username already exists")
    ErrPasswordHash      = errors.New("failed to hash password")
)

// Login 验证用户名和密码，并生成 JWT
func (a *Auth) Login(userName string, password string) (string, error) {
    dbPwd, err := a.userDao.GetPasswordByUserName(userName)
    if err != nil {
        if errors.Is(err, dao.ErrUserNotFound) {
            return "", ErrUserNotFound
        }
        return "", fmt.Errorf("database error: %v", err)
    }

    ok := bcrypt.CompareHashAndPassword([]byte(dbPwd), []byte(password))
    if ok != nil {
        return "", ErrPasswordIncorrect
    }

    // 生成 JWT
    token, err := a.generateJWT(userName)
    if err != nil {
        return "", ErrTokenGeneration
    }

    return token, nil
}

// generateJWT 生成包含 userName 的 JWT
func (a *Auth) generateJWT(userName string) (string, error) {
    claims := jwt.MapClaims{
        "userName": userName,
        "exp":      time.Now().Add(time.Hour * 24).Unix(), // 设置过期时间为 24 小时
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(a.jwtSecret)
}

// ValidateToken 验证JWT token并返回用户名
func (a *Auth) ValidateToken(tokenString string) (string, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return a.jwtSecret, nil
    })

    if err != nil {
        return "", err
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        if userName, exists := claims["userName"]; exists {
            return userName.(string), nil
        }
        return "", fmt.Errorf("userName not found in token")
    }

    return "", fmt.Errorf("invalid token")
}

// JWTAuthMiddleware JWT认证中间件
func (a *Auth) JWTAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, response.Error(response.UnauthorizedCode, "Authorization header required"))
            c.Abort()
            return
        }

        // 检查Bearer格式
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, response.Error(response.UnauthorizedCode, "Authorization header format must be Bearer {token}"))
            c.Abort()
            return
        }

        tokenString := parts[1]
        userName, err := a.ValidateToken(tokenString)
        if err != nil {
            c.JSON(http.StatusUnauthorized, response.Error(response.UnauthorizedCode, fmt.Sprintf("Invalid token: %s", err.Error())))
            c.Abort()
            return
        }

        // 将用户名存储到上下文中
        c.Set("userName", userName)
        c.Next()
    }
}

// Register 用户注册
func (a *Auth) Register(email, userName, password string) error {
    // 检查邮箱是否已存在
    emailExists, err := a.userDao.CheckUserExistsByEmail(email)
    if err != nil {
        return fmt.Errorf("database error: %v", err)
    }
    if emailExists {
        return ErrEmailExists
    }

    // 检查用户名是否已存在
    userNameExists, err := a.userDao.CheckUserExistsByUserName(userName)
    if err != nil {
        return fmt.Errorf("database error: %v", err)
    }
    if userNameExists {
        return ErrUserNameExists
    }

    // 对密码进行哈希加密
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return ErrPasswordHash
    }

    // 创建用户对象
    user := &model.User{
        Email:    email,
        UserName: userName,
        Password: string(hashedPassword),
    }

    // 保存到数据库
    err = a.userDao.CreateUser(user)
    if err != nil {
        return fmt.Errorf("failed to create user: %v", err)
    }

    return nil
}