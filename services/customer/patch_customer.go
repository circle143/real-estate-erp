package customer

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
)

type hUpdateCustomerDetails struct {
	details customerDetails
}

func (uc *hUpdateCustomerDetails) validate(db *gorm.DB, orgId, society, flatId, customerId string) error {
	customerSocietyInfo := CreateCustomerSocietyInfoService(db, uuid.MustParse(flatId), uuid.MustParse(customerId))
	err := common.IsSameSociety(customerSocietyInfo, orgId, society)
	if err != nil {
		return err
	}
	return uc.details.validate()
}

func (uc *hUpdateCustomerDetails) execute(db *gorm.DB, orgId, society, flatId, customerId string) error {
	err := uc.validate(db, orgId, society, flatId, customerId)
	if err != nil {
		return err
	}

	customerModel := models.Customer{
		Id: uuid.MustParse(customerId),
	}
	return db.Model(&customerModel).Updates(models.Customer{
		FlatId:           uuid.MustParse(flatId),
		Level:            uc.details.Level,
		Salutation:       custom.Salutation(uc.details.Salutation),
		FirstName:        uc.details.FirstName,
		LastName:         uc.details.LastName,
		DateOfBirth:      uc.details.DateOfBirth,
		Gender:           custom.Gender(uc.details.Gender),
		Photo:            uc.details.Photo,
		MaritalStatus:    custom.MaritalStatus(uc.details.MaritalStatus),
		Nationality:      custom.Nationality(uc.details.Nationality),
		Email:            uc.details.Email,
		PhoneNumber:      uc.details.PhoneNumber,
		MiddleName:       uc.details.MiddleName,
		NumberOfChildren: uc.details.NumberOfChildren,
		AnniversaryDate:  uc.details.AnniversaryDate,
		AadharNumber:     uc.details.AadharNumber,
		PanNumber:        uc.details.PanNumber,
		PassportNumber:   uc.details.PassportNumber,
		Profession:       uc.details.Profession,
		Designation:      uc.details.Designation,
		CompanyName:      uc.details.CompanyName,
	}).Error
}

func (cs *customerService) updateCustomerDetails(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	flatId := chi.URLParam(r, "flat")
	customerId := chi.URLParam(r, "customer")

	reqBody := payload.ValidateAndDecodeRequest[hUpdateCustomerDetails](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(cs.db, orgId, societyRera, flatId, customerId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully updated customer details."

	payload.EncodeJSON(w, http.StatusCreated, response)
}