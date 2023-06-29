package common

import (
	_ "github.com/mattn/go-sqlite3"
	"github.com/open_tool/app/utils/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

func InitDB(path string) {
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		//解决建表时表名自动带复数，如record变成records
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
		//打印原生sql日志
		//Logger: ormLogger.Default.LogMode(ormLogger.Info),
	})
	//设置数据库自动创建表字段（）
	if err != nil {
		logger.Error("连接数据库失败" + err.Error())
	}
	DB = db
	logger.Info("数据库连接成功")
}

func GetDB() *gorm.DB {
	return DB
}
