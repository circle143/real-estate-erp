package tower

import (
	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/thedatashed/xlsxreader"
	"gorm.io/gorm"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
)

type hCreateTower struct {
	FloorCount int    `validate:"required"`
	Name       string `validate:"required"`
}

func (h *hCreateTower) execute(db *gorm.DB, orgId, society string) (*models.Tower, error) {
	tower := models.Tower{
		OrgId:      uuid.MustParse(orgId),
		SocietyId:  society,
		FloorCount: h.FloorCount,
		Name:       h.Name,
	}

	result := db.Create(&tower)
	if result.Error != nil {
		return nil, result.Error
	}

	return &tower, nil
}

func (s *towerService) createTower(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")

	reqBody := payload.ValidateAndDecodeRequest[hCreateTower](w, r)
	if reqBody == nil {
		return
	}

	tower, err := reqBody.execute(s.db, orgId, societyRera)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully created new tower."
	response.Data = tower

	payload.EncodeJSON(w, http.StatusCreated, response)
}

type hBulkCreateTower struct{}

func (h *hBulkCreateTower) validate(r *http.Request) (multipart.File, error) {
	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Missing file in form data",
		}
	}

	// Check extension
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".xlsx") {
		return nil, &custom.RequestError{
			Status:  http.StatusUnsupportedMediaType,
			Message: "Only .xlsx files are allowed",
		}
	}

	// Optional: Check magic number (first few bytes of file)
	buffer := make([]byte, 4)
	if _, err := file.Read(buffer); err != nil {
		return nil, &custom.RequestError{
			Status:  http.StatusInternalServerError,
			Message: "Unable to read file",
		}
	}

	// Reset file read pointer to 0 for further processing later
	if _, err := file.Seek(0, 0); err != nil {
		return nil, err
	}

	return file, nil
}

func (h *hBulkCreateTower) createTowersFromFile(file multipart.File, orgId, society string, db *gorm.DB) ([]*models.Tower, error) {
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	xl, err := xlsxreader.NewReader(fileBytes)
	if err != nil {
		return nil, err
	}
	rows := xl.ReadRows(xl.Sheets[0])
	headerRow := <-rows
	columnMap := h.getColumnMap(&headerRow)

	towers := h.parseTowerRows(rows, columnMap, orgId, society)
	err = db.Create(towers).Error

	return towers, err
}

// getColumnMap maps required headers to their corresponding Excel column letters
func (h *hBulkCreateTower) getColumnMap(headerRow *xlsxreader.Row) map[string]string {
	const (
		towerNameHeader       = "Towers"
		towerFloorCountHeader = "No. Of Floors In Towers"
	)

	columnMap := make(map[string]string)

	for _, cell := range headerRow.Cells {
		val := strings.TrimSpace(cell.Value)
		switch val {
		case towerNameHeader:
			columnMap["name"] = cell.Column
		case towerFloorCountHeader:
			columnMap["floorCount"] = cell.Column
		}
	}

	if columnMap["name"] == "" || columnMap["floorCount"] == "" {
		log.Fatalf("Required columns not found in header row")
	}

	return columnMap
}

// parseTowerRows extracts tower data from Excel rows
func (h *hBulkCreateTower) parseTowerRows(rows chan xlsxreader.Row, columnMap map[string]string, orgId, society string) []*models.Tower {
	orgUUID := uuid.MustParse(orgId)
	var towers []*models.Tower

	for row := range rows {
		tower := models.Tower{
			OrgId:     orgUUID,
			SocietyId: society,
		}

		for _, cell := range row.Cells {
			switch cell.Column {
			case columnMap["name"]:
				tower.Name = strings.TrimSpace(cell.Value)
			case columnMap["floorCount"]:
				if cell.Type == xlsxreader.TypeNumerical {
					floorCount, err := strconv.Atoi(cell.Value)
					if err != nil {
						log.Fatalf("Error parsing floor count: %v", err)
					}
					tower.FloorCount = floorCount
				}
			}
		}

		// Ignore empty rows
		if tower.Name != "" {
			towers = append(towers, &tower)
		}
	}

	return towers
}

func (h *hBulkCreateTower) execute(db *gorm.DB, r *http.Request, orgId, society string) ([]*models.Tower, error) {
	file, err := h.validate(r)
	if err != nil {
		return nil, err
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)

	towers, err := h.createTowersFromFile(file, orgId, society, db)
	if err != nil {
		if strings.Contains(err.Error(), "tower_society_org_unique") {
			return nil, &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: "Duplicate Tower Name found.",
			}
		}
	}
	return towers, err
}

func (s *towerService) bulkCreateTower(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")

	err := payload.ParseMultipartForm(w, r)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	handler := hBulkCreateTower{}
	towers, err := handler.execute(s.db, r, orgId, societyRera)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully created new towers."
	response.Data = towers

	payload.EncodeJSON(w, http.StatusCreated, response)
}
