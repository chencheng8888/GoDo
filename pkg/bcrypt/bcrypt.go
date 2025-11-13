package bcrypt

import "golang.org/x/crypto/bcrypt"

// HashPassword 接收明文密码，返回 bcrypt 哈希字符串
func HashPassword(password string) (string, error) {
	// cost 越高越安全，但计算越慢；默认 10 通常够用
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPasswordHash 校验密码是否匹配哈希
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
