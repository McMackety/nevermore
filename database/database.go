package database

import (
	"github.com/McMackety/nevermore/config"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var Database gorm.DB

func InitDatabase() {
	Database, err := gorm.Open(config.DefaultConfig.Database.Type, config.DefaultConfig.Database.Type)
	if err != nil {
		panic("failed to connect database")
	}

	Database.AutoMigrate(&User{})

}
