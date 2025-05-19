package sale

import (
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type saleService struct {
	db *gorm.DB
}

func CreateSaleService(app common.IApp) common.IService {
	return &saleService{
		db: app.GetDBClient(),
	}
}

// customerDetails contains customer info for creation and update
type customerDetails struct {
	Level            int              `json:"level"`
	Salutation       string           `json:"salutation" validate:"required"`
	FirstName        string           `json:"firstName" validate:"required"`
	LastName         string           `json:"lastName" validate:"required"`
	DateOfBirth      custom.DateOnly  `json:"dateOfBirth" validate:"required"`
	Gender           string           `json:"gender" validate:"required"`
	Photo            string           `json:"photo"`
	MaritalStatus    string           `json:"maritalStatus" validate:"required"`
	Nationality      string           `json:"nationality" validate:"required"`
	Email            string           `json:"email" validate:"required,email"`
	PhoneNumber      string           `json:"phoneNumber" validate:"required,e164"`
	MiddleName       string           `json:"middleName"`
	NumberOfChildren int              `json:"numberOfChildren"`
	AnniversaryDate  *custom.DateOnly `json:"anniversaryDate"`
	AadharNumber     string           `json:"aadharNumber" validate:"omitempty,aadhar"`
	PanNumber        string           `json:"panNumber" validate:"omitempty,pan"`
	PassportNumber   string           `json:"passportNumber" validate:"omitempty,passport"`
	Profession       string           `json:"profession"`
	Designation      string           `json:"designation"`
	CompanyName      string           `json:"companyName"`
}

func (cd *customerDetails) validate() error {
	invalidError := &custom.RequestError{
		Status:  http.StatusBadRequest,
		Message: "Invalid details provided.",
	}

	// validate salutation
	salutation := custom.Salutation(cd.Salutation)
	if !salutation.IsValid() {
		return invalidError
	}

	// validate gender
	gender := custom.Gender(cd.Gender)
	if !gender.IsValid() {
		return invalidError
	}

	// validate martialStatus
	martialStatus := custom.MaritalStatus(cd.MaritalStatus)
	if !martialStatus.IsValid() {
		return invalidError
	}

	// validate nationality
	nationality := custom.Nationality(cd.Nationality)
	if !nationality.IsValid() {
		return invalidError
	}

	// validate verification info
	if strings.TrimSpace(cd.AadharNumber) == "" && strings.TrimSpace(cd.PanNumber) == "" && strings.TrimSpace(cd.PassportNumber) == "" {
		return invalidError
	}

	return nil
}
