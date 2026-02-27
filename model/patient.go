package db

import "gorm.io/gorm"

type Patient struct {
	gorm.Model
	Hospital     string `gorm:"index;uniqueIndex:idx_hospital_national_id;uniqueIndex:idx_hospital_passport_id"`
	FirstNameTh  string
	MiddleNameTh string
	LastNameTh   string
	FirstNameEn  string
	MiddleNameEn string
	LastNameEn   string
	DateOfBirth  string
	PatientHn    string
	NationalID   string `gorm:"uniqueIndex:idx_hospital_national_id"`
	PassportID   string `gorm:"uniqueIndex:idx_hospital_passport_id"`
	PhoneNumber  string
	Email        string
	Gender       string
}
