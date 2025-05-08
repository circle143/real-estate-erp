package common

import (
	"circledigital.in/real-state-erp/utils/custom"
	"github.com/google/uuid"
	"net/http"
)

type SocietyInfo struct {
	SocietyRera string
	OrgId       uuid.UUID
}

// ISocietyInfo interface is implemented by tower, flat-type and flats
type ISocietyInfo interface {
	GetSocietyInfo() (*SocietyInfo, error)
}

// IsSameSociety is a helper method to check if society details match
func IsSameSociety(societyInfoService ISocietyInfo, orgId, societyRera string) error {
	societyInfo, err := societyInfoService.GetSocietyInfo()
	if err != nil {
		return err
	}

	if societyInfo.SocietyRera != societyRera || societyInfo.OrgId.String() != orgId {
		return &custom.RequestError{
			Status:  http.StatusForbidden,
			Message: "Society mismatch.",
		}
	}
	return nil
}
