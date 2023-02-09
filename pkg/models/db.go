package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	ID       string
	Username string
	Password string
}
