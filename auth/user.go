package auth

import (
	"fmt"
	"github.com/chencheng8888/GoDo/dao"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	userDao *dao.UserDao
}

func (a *Auth) Login(userName string, password string) error {
	dbPwd, err := a.userDao.GetPasswordByUserName(userName)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	ok := bcrypt.CompareHashAndPassword([]byte(dbPwd), []byte(password))
	if ok != nil {
		return fmt.Errorf("password incorrect")
	}
	return nil
}
