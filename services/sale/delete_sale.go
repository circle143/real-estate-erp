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

type hClearSaleRecord struct{}

func (h *hClearSaleRecord) validate(db *gorm.DB, orgId, society, saleId string) error {
	societyInfoService := CreateSaleSocietyInfoService(db, uuid.MustParse(saleId))
	return common.IsSameSociety(societyInfoService, orgId, society)

}
func (h *hClearSaleRecord) execute(db *gorm.DB, orgId, society, saleId string) error {
	err := h.validate(db, orgId, society, saleId)
	if err != nil {
		return err
	}

	saleModel := models.Sale{
		Id: uuid.MustParse(saleId),
	}
	return db.Delete(&saleModel).Error
}

func (s *saleService) clearSaleRecord(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	saleId := chi.URLParam(r, "saleId")

	sale := hClearSaleRecord{}
	err := sale.execute(s.db, orgId, societyRera, saleId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully deleted sale record."

	payload.EncodeJSON(w, http.StatusOK, response)
}
