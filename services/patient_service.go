package services

import (
	"errors"
	"strings"

	db "AgnosAssignments/model"

	"gorm.io/gorm"
)

type PatientService struct {
	db *gorm.DB
}

type CreatePatientInput struct {
	Hospital     string `json:"-"`
	FirstNameTh  string `json:"first_name_th"`
	MiddleNameTh string `json:"middle_name_th"`
	LastNameTh   string `json:"last_name_th"`
	FirstNameEn  string `json:"first_name_en"`
	MiddleNameEn string `json:"middle_name_en"`
	LastNameEn   string `json:"last_name_en"`
	DateOfBirth  string `json:"date_of_birth"`
	PatientHn    string `json:"patient_hn"`
	NationalID   string `json:"national_id"`
	PassportID   string `json:"passport_id"`
	PhoneNumber  string `json:"phone_number"`
	Email        string `json:"email"`
	Gender       string `json:"gender"`
}

type SearchPatientInput struct {
	Hospital   string
	NationalID string
	PassportID string
}

func NewPatientService(database *gorm.DB) *PatientService {
	return &PatientService{db: database}
}

func (s *PatientService) CreatePatient(input CreatePatientInput) (*db.Patient, error) {
	patient := db.Patient{
		Hospital:     strings.TrimSpace(input.Hospital),
		FirstNameTh:  strings.TrimSpace(input.FirstNameTh),
		MiddleNameTh: strings.TrimSpace(input.MiddleNameTh),
		LastNameTh:   strings.TrimSpace(input.LastNameTh),
		FirstNameEn:  strings.TrimSpace(input.FirstNameEn),
		MiddleNameEn: strings.TrimSpace(input.MiddleNameEn),
		LastNameEn:   strings.TrimSpace(input.LastNameEn),
		DateOfBirth:  strings.TrimSpace(input.DateOfBirth),
		PatientHn:    strings.TrimSpace(input.PatientHn),
		NationalID:   strings.TrimSpace(input.NationalID),
		PassportID:   strings.TrimSpace(input.PassportID),
		PhoneNumber:  strings.TrimSpace(input.PhoneNumber),
		Email:        strings.TrimSpace(input.Email),
		Gender:       strings.TrimSpace(input.Gender),
	}

	if err := s.db.Create(&patient).Error; err != nil {
		return nil, err
	}
	return &patient, nil
}

func (s *PatientService) SearchPatient(input SearchPatientInput) ([]db.Patient, error) {
	nid := strings.TrimSpace(input.NationalID)
	pid := strings.TrimSpace(input.PassportID)
	if (nid == "" && pid == "") || (nid != "" && pid != "") {
		return nil, errors.New("provide exactly one of national_id or passport_id")
	}

	query := s.db.Model(&db.Patient{}).Where("hospital = ?", strings.TrimSpace(input.Hospital))

	if nid != "" {
		query = query.Where("national_id = ?", nid)
	}
	if pid != "" {
		query = query.Where("passport_id = ?", pid)
	}

	var patients []db.Patient
	if err := query.Find(&patients).Error; err != nil {
		return nil, err
	}

	return patients, nil
}
