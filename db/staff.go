package db

import "gorm.io/gorm"

type Staff struct {
	gorm.Model
	Username     string `gorm:"uniqueIndex:idx_username_hospital"`
	PasswordHash string
	Hospital     string `gorm:"index;uniqueIndex:idx_username_hospital"`
}
