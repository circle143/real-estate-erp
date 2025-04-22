package customer

import (
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
)

type customerService struct {
	db *gorm.DB
}

func CreateCustomerService(app common.IApp) common.IService {
	return &customerService{
		db: app.GetDBClient(),
	}
}

// customerDetails contains customer info for creation and update
type customerDetails struct {
	Level            int       `json:"level" validate:"required"`
	Salutation       string    `json:"salutation" validate:"required"`
	FirstName        string    `json:"firstName" validate:"required"`
	LastName         string    `json:"lastName" validate:"required"`
	DateOfBirth      time.Time `json:"dateOfBirth" validate:"required"`
	Gender           string    `json:"gender" validate:"required"`
	Photo            string    `json:"photo" validate:"required"`
	MaritalStatus    string    `json:"maritalStatus" validate:"required"`
	Nationality      string    `json:"nationality" validate:"required"`
	Email            string    `json:"email" validate:"required,email"`
	PhoneNumber      string    `json:"phoneNumber" validate:"required,e164"`
	MiddleName       string    `json:"middleName"`
	NumberOfChildren int       `json:"numberOfChildren"`
	AnniversaryDate  time.Time `json:"anniversaryDate"`
	AadharNumber     string    `json:"aadharNumber"`
	PanNumber        string    `json:"panNumber"`
	PassportNumber   string    `json:"passportNumber"`
	Profession       string    `json:"profession"`
	Designation      string    `json:"designation"`
	CompanyName      string    `json:"companyName"`
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