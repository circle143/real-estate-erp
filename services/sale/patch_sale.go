package sale

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

type hUpdateCustomerSaleDetails struct {
	Details customerDetails `validate:"required"`
}

func (h *hUpdateCustomerSaleDetails) validate(db *gorm.DB, orgId, society, customerId string) error {
	customerSocietyInfo := CreateCustomerSocietyInfoService(db, uuid.MustParse(customerId))
	err := common.IsSameSociety(customerSocietyInfo, orgId, society)
	if err != nil {
		return err
	}
	return h.Details.validate()
}

func (h *hUpdateCustomerSaleDetails) execute(db *gorm.DB, orgId, society, customerId string) error {
	err := h.validate(db, orgId, society, customerId)
	if err != nil {
		return err
	}

	customerModel := models.Customer{
		Id: uuid.MustParse(customerId),
	}
	return db.Model(&customerModel).Updates(models.Customer{
		Salutation:       custom.Salutation(h.Details.Salutation),
		FirstName:        h.Details.FirstName,
		LastName:         h.Details.LastName,
		DateOfBirth:      h.Details.DateOfBirth,
		Gender:           custom.Gender(h.Details.Gender),
		Photo:            h.Details.Photo,
		MaritalStatus:    custom.MaritalStatus(h.Details.MaritalStatus),
		Nationality:      custom.Nationality(h.Details.Nationality),
		Email:            h.Details.Email,
		PhoneNumber:      h.Details.PhoneNumber,
		MiddleName:       h.Details.MiddleName,
		NumberOfChildren: h.Details.NumberOfChildren,
		AnniversaryDate:  h.Details.AnniversaryDate,
		AadharNumber:     h.Details.AadharNumber,
		PanNumber:        h.Details.PanNumber,
		PassportNumber:   h.Details.PassportNumber,
		Profession:       h.Details.Profession,
		Designation:      h.Details.Designation,
		CompanyName:      h.Details.CompanyName,
	}).Error
}

func (s *saleService) updateSaleCustomerDetails(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	modelId := chi.URLParam(r, "customerId")

	reqBody := payload.ValidateAndDecodeRequest[hUpdateCustomerSaleDetails](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(s.db, orgId, societyRera, modelId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully updated customer details."

	payload.EncodeJSON(w, http.StatusOK, response)
}

type hUpdateCompanyCustomerSaleDetails struct {
	Details companyCustomerDetails `validate:"required"`
}

func (h *hUpdateCompanyCustomerSaleDetails) validate(db *gorm.DB, orgId, society, customerId string) error {
	companyCustomerSocietyInfo := CreateCompanyCustomerSocietyInfoService(db, uuid.MustParse(customerId))
	err := common.IsSameSociety(companyCustomerSocietyInfo, orgId, society)
	if err != nil {
		return err
	}
	return h.Details.validate()
}

func (h *hUpdateCompanyCustomerSaleDetails) execute(db *gorm.DB, orgId, society, customerId string) error {
	err := h.validate(db, orgId, society, customerId)
	if err != nil {
		return err
	}

	customerModel := models.CompanyCustomer{
		Id: uuid.MustParse(customerId),
	}
	return db.Model(&customerModel).Updates(models.CompanyCustomer{
		Name:         h.Details.Name,
		CompanyPan:   h.Details.CompanyPan,
		CompanyGst:   h.Details.CompanyGst,
		AadharNumber: h.Details.AadharNumber,
		PanNumber:    h.Details.PanNumber,
	}).Error
}

func (s *saleService) updateSaleCompanyCustomerDetails(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	modelId := chi.URLParam(r, "customerId")

	reqBody := payload.ValidateAndDecodeRequest[hUpdateCompanyCustomerSaleDetails](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(s.db, orgId, societyRera, modelId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully updated customer details."

	payload.EncodeJSON(w, http.StatusOK, response)
}
