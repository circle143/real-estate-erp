package customer

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/services/flat"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
)

type hAddCustomerToFlat struct {
	Details []customerDetails `validate:"required,min=1,dive"`
	Seller  string            `validate:"required"`
}

func (ac *hAddCustomerToFlat) validate(db *gorm.DB, orgId, society, flatId string) error {
	societyInfoService := flat.CreateFlatSocietyInfoService(db, uuid.MustParse(flatId))
	err := common.IsSameSociety(societyInfoService, orgId, society)
	if err != nil {
		return err
	}

	seller := custom.Seller(ac.Seller)
	if !seller.IsValid() {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid seller value provided.",
		}
	}

	for _, detail := range ac.Details {
		err = detail.validate()
		if err != nil {
			return err
		}
	}

	return nil
}

func (ac *hAddCustomerToFlat) execute(db *gorm.DB, orgId, society, flatId string) error {
	err := ac.validate(db, orgId, society, flatId)
	if err != nil {
		return err
	}

	customers := make([]*models.Customer, 0, len(ac.Details))

	for _, d := range ac.Details {
		customer := &models.Customer{
			FlatId:           uuid.MustParse(flatId),
			Level:            d.Level,
			Salutation:       custom.Salutation(d.Salutation),
			FirstName:        d.FirstName,
			LastName:         d.LastName,
			DateOfBirth:      d.DateOfBirth,
			Gender:           custom.Gender(d.Gender),
			Photo:            d.Photo,
			MaritalStatus:    custom.MaritalStatus(d.MaritalStatus),
			Nationality:      custom.Nationality(d.Nationality),
			Email:            d.Email,
			PhoneNumber:      d.PhoneNumber,
			MiddleName:       d.MiddleName,
			NumberOfChildren: d.NumberOfChildren,
			AnniversaryDate:  d.AnniversaryDate,
			AadharNumber:     d.AadharNumber,
			PanNumber:        d.PanNumber,
			PassportNumber:   d.PassportNumber,
			Profession:       d.Profession,
			Designation:      d.Designation,
			CompanyName:      d.CompanyName,
		}
		customers = append(customers, customer)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(customers).Error
		if err != nil {
			return err
		}

		// update flat to sold
		flatModel := models.Flat{
			Id: uuid.MustParse(flatId),
		}
		return tx.Model(&flatModel).Update("sold_by", ac.Seller).Error

	})
}

func (cs *customerService) addCustomerToFlat(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	flatId := chi.URLParam(r, "flat")

	reqBody := payload.ValidateAndDecodeRequest[hAddCustomerToFlat](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(cs.db, orgId, societyRera, flatId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully added customer to flats."

	payload.EncodeJSON(w, http.StatusCreated, response)
}
