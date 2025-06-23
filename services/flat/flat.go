package flat

import (
	"circledigital.in/real-state-erp/utils/common"
	"circledigital.in/real-state-erp/utils/custom"
	"gorm.io/gorm"
	"net/http"
	"strconv"
	"strings"
	"unicode"
)

type flatService struct {
	db *gorm.DB
}

func CreateFlatService(app common.IApp) common.IService {
	return &flatService{
		db: app.GetDBClient(),
	}
}

// parseFlatIdentifier parses identifiers like "asd-1001" or "A-101"
func parseFlatIdentifier(input string) (int, error) {
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
