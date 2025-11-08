package dao

import (
	"github.com/chencheng8888/GoDo/config"
	"github.com/chencheng8888/GoDo/model"
	"github.com/google/wire"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"moul.io/zapgorm2"
)

var (
	ProviderSet = wire.NewSet(NewDB, NewTaskLogDao)
)

func NewDB(cf *config.DBConfig, log *zap.SugaredLogger) (*gorm.DB, error) {
	logger := zapgorm2.New(log.Desugar())
	db, err := gorm.Open(mysql.Open(cf.Addr), &gorm.Config{
		Logger: logger,
	})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&model.TaskLog{})
	if err != nil {
		return nil, err
	}
	log.Infof("✅ database connected successfully")
	return db, nil
}
