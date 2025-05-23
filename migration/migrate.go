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
		&models.PriceHistory{},
		&models.PreferenceLocationCharge{},
		&models.OtherCharge{},
		&models.Sale{},
		&models.PaymentPlan{},
		&models.TowerPaymentStatus{},
		&models.SalePaymentStatus{},
		&models.CompanyCustomer{},
		&models.Broker{},
		&models.Bank{},
	)

	//err := db.Migrator().DropTable(
	//	//&models.Organization{},
	//	//&models.User{},
	//	//&models.Society{},
	//	//&models.FlatType{},
	//	//&models.Tower{},
	//	//&models.Flat{},
	//	//&models.Customer{},
	//	//&models.Sale{},
	//	//&models.CompanyCustomer{},
	//	&models.SalePaymentStatus{},
	//)

	if err != nil {
		log.Fatalf("Error migrating db: %v", err)
	}
	log.Println("Successfully migrated database models.")
}
