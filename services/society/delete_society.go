package society

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"net/http"
)

type hDeleteSociety struct{}

func (h *hDeleteSociety) execute(db *gorm.DB, society, orgId string) error {
	societyModel := models.Society{
		ReraNumber: society,
		OrgId:      uuid.MustParse(orgId),
	}

	err := db.Delete(&societyModel).Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: "You need to first delete society resources to delete society.",
			}
		}
		return err
	}
	return nil
}

func (s *societyService) deleteSociety(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")

	society := hDeleteSociety{}
	err := society.execute(s.db, societyRera, orgId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully deleted society."

	payload.EncodeJSON(w, http.StatusOK, response)
}
