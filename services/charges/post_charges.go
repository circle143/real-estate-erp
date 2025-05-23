package charges

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"net/http"
)

type hAddNewPreferenceLocationCharge struct {
	Summary string `validate:"required"`
	Type    string `validate:"required"`
	Floor   int
	Price   float64
}

func (h *hAddNewPreferenceLocationCharge) validate() error {
	chargeType := custom.PreferenceLocationChargesType(h.Type)
	if !chargeType.IsValid() {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid preference location charge type.",
		}
	}
	return nil
}

func (h *hAddNewPreferenceLocationCharge) execute(db *gorm.DB, orgId, society string) (*models.PreferenceLocationCharge, error) {
	err := h.validate()
	if err != nil {
		return nil, err
	}

	chargeModel := models.PreferenceLocationCharge{
		OrgId:     uuid.MustParse(orgId),
		SocietyId: society,
		Summary:   h.Summary,
		Type:      custom.PreferenceLocationChargesType(h.Type),
		Floor:     h.Floor,
		Price:     decimal.NewFromFloat(h.Price),
	}

	// transaction to create preference location charge and update in price table
	err = db.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&chargeModel).Error
		if err != nil {
			return nil
		}

		priceUtil := createPriceUtil(tx, chargeModel.Id, custom.PREFERENCELOCATIONCHARGE, chargeModel.Price)
		return priceUtil.addInitialPrice()
	})
	return &chargeModel, err
}

func (s *chargesService) addNewPreferenceLocationCharge(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")

	reqBody := payload.ValidateAndDecodeRequest[hAddNewPreferenceLocationCharge](w, r)
	if reqBody == nil {
		return
	}

	charge, err := reqBody.execute(s.db, orgId, societyRera)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully added new charge."
	response.Data = charge

	payload.EncodeJSON(w, http.StatusCreated, response)
}

type hAddNewOtherCharge struct {
	Summary       string `validate:"required"`
	Recurring     bool
	Optional      bool
	Fixed         bool
	AdvanceMonths int
	Price         float64
}

func (h *hAddNewOtherCharge) validate() error {
	if (h.Recurring && h.AdvanceMonths >= 1) || (!h.Recurring && h.AdvanceMonths == 0) {
		return nil
	}

	return &custom.RequestError{
		Status:  http.StatusBadRequest,
		Message: "Invalid value for field advanceMonths",
	}
}

func (h *hAddNewOtherCharge) execute(db *gorm.DB, orgId, society string) (*models.OtherCharge, error) {
	err := h.validate()
	if err != nil {
		return nil, err
	}

	chargeModel := models.OtherCharge{
		OrgId:         uuid.MustParse(orgId),
		SocietyId:     society,
		Summary:       h.Summary,
		Recurring:     h.Recurring,
		Optional:      h.Optional,
		AdvanceMonths: h.AdvanceMonths,
		Price:         decimal.NewFromFloat(h.Price),
		Fixed:         h.Fixed,
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&chargeModel).Error
		if err != nil {
			return nil
		}

		priceUtil := createPriceUtil(tx, chargeModel.Id, custom.OTHERCHARGE, chargeModel.Price)
		return priceUtil.addInitialPrice()
	})
	return &chargeModel, err
}

func (s *chargesService) addNewOtherCharge(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")

	reqBody := payload.ValidateAndDecodeRequest[hAddNewOtherCharge](w, r)
	if reqBody == nil {
		return
	}

	charge, err := reqBody.execute(s.db, orgId, societyRera)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully added new charge."
	response.Data = charge

	payload.EncodeJSON(w, http.StatusCreated, response)
}
