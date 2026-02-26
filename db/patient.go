package db

import "gorm.io/gorm"

type Patient struct {
	gorm.Model
	Hospital     string `gorm:"index"`
	FirstNameTh  string
	MiddleNameTh string
	LastNameTh   string
	FirstNameEn  string
	MiddleNameEn string
	LastNameEn   string
	DateOfBirth  string // Use time.Time for better handling
	PatientHn    string
	NationalID   string `gorm:"uniqueIndex:idx_hospital_national"`
	PassportID   string `gorm:"uniqueIndex:idx_hospital_passport"`
	PhoneNumber  string
	Email        string
	Gender       string
}
