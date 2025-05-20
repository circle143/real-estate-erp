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
	details customerDetails `validate:"required"`
}

func (h *hUpdateCustomerSaleDetails) validate(db *gorm.DB, orgId, society, customerId string) error {
	customerSocietyInfo := CreateCustomerSocietyInfoService(db, uuid.MustParse(customerId))
	err := common.IsSameSociety(customerSocietyInfo, orgId, society)
	if err != nil {
		return err
	}
	return h.details.validate()
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
		Salutation:       custom.Salutation(h.details.Salutation),
		FirstName:        h.details.FirstName,
		LastName:         h.details.LastName,
		DateOfBirth:      h.details.DateOfBirth,
		Gender:           custom.Gender(h.details.Gender),
		Photo:            h.details.Photo,
		MaritalStatus:    custom.MaritalStatus(h.details.MaritalStatus),
		Nationality:      custom.Nationality(h.details.Nationality),
		Email:            h.details.Email,
		PhoneNumber:      h.details.PhoneNumber,
		MiddleName:       h.details.MiddleName,
		NumberOfChildren: h.details.NumberOfChildren,
		AnniversaryDate:  h.details.AnniversaryDate,
		AadharNumber:     h.details.AadharNumber,
		PanNumber:        h.details.PanNumber,
		PassportNumber:   h.details.PassportNumber,
		Profession:       h.details.Profession,
		Designation:      h.details.Designation,
		CompanyName:      h.details.CompanyName,
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

	payload.EncodeJSON(w, http.StatusCreated, response)
}

type hUpdateCompanyCustomerSaleDetails struct {
	details companyCustomerDetails `validate:"required"`
}

func (h *hUpdateCompanyCustomerSaleDetails) validate(db *gorm.DB, orgId, society, customerId string) error {
	companyCustomerSocietyInfo := CreateCompanyCustomerSocietyInfoService(db, uuid.MustParse(customerId))
	err := common.IsSameSociety(companyCustomerSocietyInfo, orgId, society)
	if err != nil {
		return err
	}
	return h.details.validate()
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
		Name:         h.details.Name,
		CompanyPan:   h.details.CompanyPan,
		CompanyGst:   h.details.CompanyGst,
		AadharNumber: h.details.AadharNumber,
		PanNumber:    h.details.PanNumber,
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

	payload.EncodeJSON(w, http.StatusCreated, response)
}
