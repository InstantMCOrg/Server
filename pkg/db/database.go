package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Init() {
	_, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}

}
