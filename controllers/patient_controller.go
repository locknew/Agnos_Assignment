package controllers

import (
	"net/http"
	"strings"

	"AgnosAssignments/services"

	"github.com/gin-gonic/gin"
)

type PatientController struct {
	patientService *services.PatientService
}

func NewPatientController(patientService *services.PatientService) *PatientController {
	return &PatientController{patientService: patientService}
}

func (ctl *PatientController) CreatePatient(c *gin.Context) {
	var input services.CreatePatientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hospital, _ := c.Get("hospital")
	input.Hospital, _ = hospital.(string)

	patient, err := ctl.patientService.CreatePatient(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": patient})
}

func (ctl *PatientController) SearchPatient(c *gin.Context) {
	hospital, _ := c.Get("hospital")
	hospitalName, _ := hospital.(string)
	nationalID := c.Query("national_id")
	passportID := c.Query("passport_id")
	patientHN := c.Query("patient_hn")

	if strings.TrimSpace(patientHN) != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "search by patient_hn is not allowed, use national_id or passport_id"})
		return
	}
	if (strings.TrimSpace(nationalID) == "" && strings.TrimSpace(passportID) == "") ||
		(strings.TrimSpace(nationalID) != "" && strings.TrimSpace(passportID) != "") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provide exactly one query: national_id or passport_id"})
		return
	}

	input := services.SearchPatientInput{
		Hospital:   hospitalName,
		NationalID: nationalID,
		PassportID: passportID,
	}

	patients, err := ctl.patientService.SearchPatient(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": patients})
}
