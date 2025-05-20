package sale

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/services/flat"
	"circledigital.in/real-state-erp/utils/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// get sale society info
type saleSocietyInfoService struct {
	db     *gorm.DB
	saleId uuid.UUID
}

func (s *saleSocietyInfoService) GetSocietyInfo() (*common.SocietyInfo, error) {
	sale := models.Sale{
		Id: s.saleId,
	}

	err := s.db.First(&sale).Error
	if err != nil {
		return nil, err
	}

	flatSocietyInfo := flat.CreateFlatSocietyInfoService(s.db, sale.FlatId)
	return flatSocietyInfo.GetSocietyInfo()
}

// get customer society info
type customerSocietyInfoService struct {
	db *gorm.DB
	id uuid.UUID
}

func (s *customerSocietyInfoService) GetSocietyInfo() (*common.SocietyInfo, error) {
	customer := models.Customer{
		Id: s.id,
	}

	err := s.db.First(&customer).Error
	if err != nil {
		return nil, err
	}

	saleSocietyInfo := CreateSaleSocietyInfoService(s.db, customer.SaleId)
	return saleSocietyInfo.GetSocietyInfo()
}

// get company customer society info
type companyCustomerSocietyInfoService struct {
	db *gorm.DB
	id uuid.UUID
}

func (s *companyCustomerSocietyInfoService) GetSocietyInfo() (*common.SocietyInfo, error) {
	customer := models.CompanyCustomer{
		Id: s.id,
	}

	err := s.db.First(&customer).Error
	if err != nil {
		return nil, err
	}

	saleSocietyInfo := CreateSaleSocietyInfoService(s.db, customer.SaleId)
	return saleSocietyInfo.GetSocietyInfo()
}

func CreateSaleSocietyInfoService(db *gorm.DB, saleId uuid.UUID) common.ISocietyInfo {
	return &saleSocietyInfoService{
		db:     db,
		saleId: saleId,
	}
}

func CreateCustomerSocietyInfoService(db *gorm.DB, id uuid.UUID) common.ISocietyInfo {
	return &customerSocietyInfoService{
		db: db,
		id: id,
	}
}

func CreateCompanyCustomerSocietyInfoService(db *gorm.DB, id uuid.UUID) common.ISocietyInfo {
	return &companyCustomerSocietyInfoService{
		db: db,
		id: id,
	}
}
