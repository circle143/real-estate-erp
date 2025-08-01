package migration

import (
	"log"

	"circledigital.in/real-state-erp/models"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) {
	err := db.AutoMigrate(
		&models.Organization{},
		&models.User{},
		&models.Society{},
		//&models.FlatType{},
		&models.Tower{},
		&models.Flat{},
		&models.Customer{},
		&models.PriceHistory{},
		&models.PreferenceLocationCharge{},
		&models.OtherCharge{},
		&models.Sale{},
		// &models.PaymentPlan{},
		&models.PaymentPlanGroup{},
		&models.PaymentPlanRatio{},
		&models.PaymentPlanRatioItem{},
		&models.TowerPaymentStatus{},
		&models.FlatPaymentStatus{},
		&models.CompanyCustomer{},
		&models.Broker{},
		&models.Bank{},
		&models.Receipt{},
		&models.ReceiptClear{},
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
	//	//&models.SalePaymentStatus{},
	//	&models.Receipt{},
	//	&models.ReceiptClear{},
	//)

	if err != nil {
		log.Fatalf("Error migrating db: %v", err)
	}
	log.Println("Successfully migrated database models.")
}
