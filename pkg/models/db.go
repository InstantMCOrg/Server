package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string
	Password string
}

type Session struct {
	gorm.Model
	Token  string
	UserID int
	User   User
}

type DBMcServerContainer struct {
	gorm.Model
	UserID int
	McServerContainer
}
