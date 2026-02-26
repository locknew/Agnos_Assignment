package db

import "gorm.io/gorm"

type Hospital struct {
	gorm.Model
	Name string `gorm:"uniqueIndex;not null"`
}
