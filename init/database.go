package init

import (
	"circledigital.in/real-state-erp/migration"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"os"
)

func (a *app) createDBClient() *gorm.DB {
	dsn := os.Getenv("DB_DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatalf("error connection to db: %v", err)
	}

	log.Println("DB connected")
	if os.Getenv("ENV") != "development" {
		migration.Migrate(db)
	}

	return db
}