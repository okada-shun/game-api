package api

import (
	"io/ioutil"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	_ "github.com/go-sql-driver/mysql"
)

// DB(game_user)からコネクション取得
func GetConnection(passwordfile string, userfile string) (*gorm.DB, error) {
	passwordBytes, err := ioutil.ReadFile(passwordfile)
	if err != nil {
		return nil, err
	}
	userBytes, err := ioutil.ReadFile(userfile)
	if err != nil {
		return nil, err
	}
	db, err := gorm.Open(mysql.Open(string(userBytes)+":"+string(passwordBytes)+"@/game_user?charset=utf8&parseTime=True&loc=Local"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.Logger = db.Logger.LogMode(logger.Info)
	return db, nil
}