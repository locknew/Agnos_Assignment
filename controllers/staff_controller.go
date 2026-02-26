package controllers

import (
	"net/http"

	"AgnosAssignments/services"

	"github.com/gin-gonic/gin"
)

type StaffController struct {
	authService *services.AuthService
}

func NewStaffController(authService *services.AuthService) *StaffController {
	return &StaffController{authService: authService}
}

func (ctl *StaffController) CreateStaff(c *gin.Context) {
	var input services.CreateStaffInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	staff, err := ctl.authService.CreateStaff(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":       staff.ID,
		"username": staff.Username,
		"hospital": staff.Hospital,
	})
}

func (ctl *StaffController) LoginStaff(c *gin.Context) {
	var input services.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := ctl.authService.Login(input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
