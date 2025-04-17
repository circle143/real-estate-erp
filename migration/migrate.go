package migration

import (
	"circledigital.in/real-state-erp/models"
	"gorm.io/gorm"
	"log"
)

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.Organization{},
		&models.User{},
		&models.Society{},
		&models.FlatType{},
		&models.Tower{},
		&models.Flat{},
		&models.Customer{},
	)

	if err != nil {
		log.Fatalf("Error migrating db: %v", err)
	}
	log.Println("Successfully migrated database models.")
}