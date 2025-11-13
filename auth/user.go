package auth

import (
    "fmt"
    "time"

    "github.com/chencheng8888/GoDo/dao"
    "github.com/golang-jwt/jwt/v4"
    "github.com/google/wire"
    "golang.org/x/crypto/bcrypt"
)

var (
    ProviderSet = wire.NewSet(NewAuth)
)

type Auth struct {
    userDao *dao.UserDao
}

var jwtSecret = []byte("your_secret_key") // 替换为更安全的密钥

// NewAuth 创建Auth实例
func NewAuth(userDao *dao.UserDao) *Auth {
    return &Auth{
        userDao: userDao,
    }
}

// Login 验证用户名和密码，并生成 JWT
func (a *Auth) Login(userName string, password string) (string, error) {
    dbPwd, err := a.userDao.GetPasswordByUserName(userName)
    if err != nil {
        return "", fmt.Errorf("user not found")
    }

    ok := bcrypt.CompareHashAndPassword([]byte(dbPwd), []byte(password))
    if ok != nil {
        return "", fmt.Errorf("password incorrect")
    }

    // 生成 JWT
    token, err := a.generateJWT(userName)
    if err != nil {
        return "", fmt.Errorf("failed to generate token: %v", err)
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
    return token.SignedString(jwtSecret)
}