package society

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type hUpdateSocietyDetails struct {
	ReraNumber string
	Name       string
	Address    string
	CoverPhoto string
}

func (usd *hUpdateSocietyDetails) validate() error {
	if strings.TrimSpace(usd.Name) == "" && strings.TrimSpace(usd.CoverPhoto) == "" && strings.TrimSpace(usd.Address) == "" && strings.TrimSpace(usd.ReraNumber) == "" {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid field values to update.",
		}
	}
	return nil
}

func (usd *hUpdateSocietyDetails) execute(db *gorm.DB, society, orgId string) error {
	err := usd.validate()
	if err != nil {
		return err
	}

	societyModel := models.Society{
		ReraNumber: society,
		OrgId:      uuid.MustParse(orgId),
	}

	return db.Model(&societyModel).Updates(models.Society{
		ReraNumber: usd.ReraNumber,
		Name:       usd.Name,
		Address:    usd.Address,
		CoverPhoto: usd.CoverPhoto,
	}).Error
}

func (s *societyService) updateSocietyDetails(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")

	reqBody := payload.ValidateAndDecodeRequest[hUpdateSocietyDetails](w, r)
	if reqBody == nil {
		return
	}

	err := reqBody.execute(s.db, societyRera, orgId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully updated society."

	payload.EncodeJSON(w, http.StatusOK, response)
}
