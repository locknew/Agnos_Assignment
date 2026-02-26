package main

import (
	"AgnosAssignments/config"
	"AgnosAssignments/controllers"
	"AgnosAssignments/middlewares"
	dbmodel "AgnosAssignments/model"
	"AgnosAssignments/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB handle: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}
	err = db.AutoMigrate(
		&dbmodel.Hospital{},
		&dbmodel.Staff{},
		&dbmodel.Patient{},
	)
	if err != nil {
		log.Fatalf("failed to migrate: %v", err)
	}
	log.Println("database connection successful")

	r := gin.Default()
	authService := services.NewAuthService(db, cfg.JWTSecret)
	patientService := services.NewPatientService(db)
	staffController := controllers.NewStaffController(authService)
	patientController := controllers.NewPatientController(patientService)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.POST("/staff/create", staffController.CreateStaff)
	r.POST("/staff/login", staffController.LoginStaff)

	protected := r.Group("/")
	protected.Use(middlewares.AuthMiddleware(authService))
	protected.POST("/patient/create", patientController.CreatePatient)
	protected.GET("/patient/search", patientController.SearchPatient)

	if err := r.Run(":8081"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

}
