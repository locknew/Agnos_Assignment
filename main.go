package main

import (
	"AgnosAssignments/config"
	dbmodel "AgnosAssignments/db"
	"log"

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

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

}
