package society

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
)

type hCreateSociety struct {
	ReraNumber string `validate:"required"`
	Name       string `validate:"required"`
	Address    string `validate:"required"`
	CoverPhoto string
}

func (h *hCreateSociety) execute(db *gorm.DB, orgId string) (*models.Society, error) {
	society := models.Society{
		ReraNumber: h.ReraNumber,
		OrgId:      uuid.MustParse(orgId),
		Name:       h.Name,
		Address:    h.Address,
		CoverPhoto: h.CoverPhoto,
	}

	result := db.Create(&society)
	if result.Error != nil {
		return nil, result.Error
	}

	return &society, nil
}

func (s *societyService) createSociety(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	reqBody := payload.ValidateAndDecodeRequest[hCreateSociety](w, r)
	if reqBody == nil {
		return
	}

	society, err := reqBody.execute(s.db, orgId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully created new society."
	response.Data = society

	payload.EncodeJSON(w, http.StatusCreated, response)
}
