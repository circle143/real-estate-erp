package flat

import (
	"circledigital.in/real-state-erp/models"
	flatType "circledigital.in/real-state-erp/services/flat-type"
	"circledigital.in/real-state-erp/services/tower"
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"fmt"
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
	"unicode"
)

type hCreateFlat struct {
	Tower       string `validate:"required,uuid"`
	FlatType    string `validate:"required,uuid"`
	Name        string `validate:"required"`
	FloorNumber int    `validate:"gte=0"`
	Facing      string `validate:"required"`
}

func (h *hCreateFlat) validate(db *gorm.DB, orgId, society string) error {
	// validate facing
	facing := custom.Facing(h.Facing)
	if !facing.IsValid() {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid flat facing value.",
		}
	}

	// validate correct flat type
	flatTypeSocietyInfo := flatType.CreateFlatTypeSocietyInfoService(db, uuid.MustParse(h.FlatType))
	err := common.IsSameSociety(flatTypeSocietyInfo, orgId, society)
	if err != nil {
		return err
	}

	// validate tower belongs to correct society and organization
	towerSocietyInfoService := tower.CreateTowerSocietyInfoService(db, uuid.MustParse(h.Tower))
	err = common.IsSameSociety(towerSocietyInfoService, orgId, society)
	if err != nil {
		return err
	}

	var towerModel models.Tower
	err = db.Where(&models.Tower{
		Id:        uuid.MustParse(h.Tower),
		OrgId:     uuid.MustParse(orgId),
		SocietyId: society,
	}).First(&towerModel).Error
	if err != nil {
		return err
	}

	// validate floor number
	if h.FloorNumber > towerModel.FloorCount {
		return &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: "Invalid floor number.",
		}
	}

	return nil
}

func (h *hCreateFlat) execute(db *gorm.DB, orgId, society string) (*models.Flat, error) {
	err := h.validate(db, orgId, society)
	if err != nil {
		return nil, err
	}

	flat := models.Flat{
		TowerId:     uuid.MustParse(h.Tower),
		FlatTypeId:  uuid.MustParse(h.FlatType),
		Name:        h.Name,
		FloorNumber: h.FloorNumber,
		Facing:      custom.Facing(h.Facing),
	}

	result := db.Create(&flat)
	if result.Error != nil {
		return nil, result.Error
	}

	return &flat, nil

}

func (s *flatService) createNewFlat(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")

	reqBody := payload.ValidateAndDecodeRequest[hCreateFlat](w, r)
	if reqBody == nil {
		return
	}

	flat, err := reqBody.execute(s.db, orgId, societyRera)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully created new flat."
	response.Data = flat

	payload.EncodeJSON(w, http.StatusCreated, response)
}

type hBulkCreateFlats struct{}

// parseFlatIdentifier parses identifiers like "asd-1001" or "A-101"
func (h *hBulkCreateFlats) parseFlatIdentifier(input string) (int, error) {
	customError := &custom.RequestError{
		Status:  http.StatusBadRequest,
		Message: "invalid format: expected something like 'A-101' or 'A-1001'",
	}

	parts := strings.Split(input, "-")
	if len(parts) != 2 {
		return -1, customError
	}

	numPart := strings.TrimSpace(parts[1])
	numPart = strings.TrimLeftFunc(numPart, func(r rune) bool {
		return !unicode.IsDigit(r)
	})

	if len(numPart) < 2 {
		return -1, customError
	}

	// Split floor and flat: floor = all digits except last 2, flat = last 2 digits
	floorStr := numPart[:len(numPart)-2]

	floor, err := strconv.Atoi(floorStr)
	if err != nil {
		return -1, customError
	}

	return floor, nil
}

func (h *hBulkCreateFlats) validate(r *http.Request) (multipart.File, error) {
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

	// Optional: Check magic number (first few bytes of a file)
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

func (h *hBulkCreateFlats) getFlatsDataFromFile(file multipart.File, orgId, society, towerId string, db *gorm.DB) ([]*models.Flat, error) {
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
	columnMap, err := h.getColumnMap(&headerRow)
	if err != nil {
		return nil, err
	}

	// get all flat types
	var flatTypes []models.FlatType
	tx := db.Where("org_id = ? and society_id = ?", orgId, society).Find(&flatTypes)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 {
		return nil, &custom.RequestError{
			Status:  http.StatusNotFound,
			Message: "You need to first create society flat types to bulk create flats.",
		}
	}

	flats, err := h.parseFlatRows(rows, columnMap, towerId, flatTypes)
	if err != nil {
		return nil, err
	}

	err = db.Create(flats).Error
	return flats, nil
}

// getColumnMap maps required headers to their corresponding Excel column letters
func (h *hBulkCreateFlats) getColumnMap(headerRow *xlsxreader.Row) (map[string]string, error) {
	const (
		nameHeader      = "Unit No"
		superAreaHeader = "Saleable Area (in sq. ft.)"
		facing          = "Flat facing"
	)

	columnMap := make(map[string]string)

	for _, cell := range headerRow.Cells {
		val := strings.TrimSpace(cell.Value)
		switch val {
		case nameHeader:
			columnMap["name"] = cell.Column
		case superAreaHeader:
			columnMap["area"] = cell.Column
		case facing:
			columnMap["facing"] = cell.Column
		}
	}

	if columnMap["name"] == "" || columnMap["area"] == "" || columnMap["facing"] == "" {
		return nil, &custom.RequestError{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("Required columns ('%v', '%v' and '%v') not found.", nameHeader, superAreaHeader, facing),
		}
	}

	return columnMap, nil
}

// parseFlatRows extracts tower data from Excel rows
func (h *hBulkCreateFlats) parseFlatRows(rows chan xlsxreader.Row, columnMap map[string]string, towerId string, flatTypes []models.FlatType) ([]*models.Flat, error) {
	flatTypeMap := make(map[string]uuid.UUID)
	for _, ft := range flatTypes {
		flatTypeMap[ft.SuperArea.String()] = ft.Id
		log.Println(ft.SuperArea)
	}

	towerUUID := uuid.MustParse(towerId)
	var flats []*models.Flat

	for row := range rows {
		flat := models.Flat{
			TowerId: towerUUID,
		}

		for _, cell := range row.Cells {
			switch cell.Column {
			case columnMap["name"]:
				flat.Name = strings.TrimSpace(cell.Value)
				floor, err := h.parseFlatIdentifier(flat.Name)
				if err != nil {
					return nil, &custom.RequestError{
						Status:  http.StatusBadRequest,
						Message: fmt.Sprintf("Invalid flat unit number provided: %v", cell.Value),
					}
				}
				flat.FloorNumber = floor
			case columnMap["area"]:
				if cell.Type == xlsxreader.TypeNumerical {
					if _, ok := flatTypeMap[cell.Value]; !ok {
						return nil, &custom.RequestError{
							Status:  http.StatusBadRequest,
							Message: fmt.Sprintf("No flat type created for super area: %v", cell.Value),
						}
					}
					flat.FlatTypeId = flatTypeMap[cell.Value]
				}
			case columnMap["facing"]:
				facing := custom.Facing(cell.Value)
				if !facing.IsValid() {
					return nil, &custom.RequestError{
						Status:  http.StatusBadRequest,
						Message: fmt.Sprintf("Invalid flat facing value provided: %v\nRequired values are: '%v' and '%v'", cell.Value, custom.DEFAULT, custom.SPECIAL),
					}
				}
				flat.Facing = facing
			}
		}

		// Ignore empty rows
		if flat.Name != "" {
			flats = append(flats, &flat)
		}
	}

	return flats, nil
}

func (h *hBulkCreateFlats) execute(db *gorm.DB, r *http.Request, orgId, society, towerId string) ([]*models.Flat, error) {
	file, err := h.validate(r)
	if err != nil {
		return nil, err
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)

	flats, err := h.getFlatsDataFromFile(file, orgId, society, towerId, db)
	if err != nil {
		if strings.Contains(err.Error(), "tower_flat_unique") {
			return nil, &custom.RequestError{
				Status:  http.StatusBadRequest,
				Message: "Duplicate flat name found.",
			}
		}
	}
	return flats, err
}

func (s *flatService) createBulkFlats(w http.ResponseWriter, r *http.Request) {
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")
	towerId := chi.URLParam(r, "towerId")

	err := payload.ParseMultipartForm(w, r)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	handler := hBulkCreateFlats{}
	flats, err := handler.execute(s.db, r, orgId, societyRera, towerId)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	var response custom.JSONResponse
	response.Error = false
	response.Message = "Successfully created new flats."
	response.Data = flats

	payload.EncodeJSON(w, http.StatusCreated, response)
}
