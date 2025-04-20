package common

import "github.com/google/uuid"

type SocietyInfo struct {
	SocietyRera string
	OrgId       uuid.UUID
}

// ISocietyInfo interface is implemented by tower, flat-type and flats
type ISocietyInfo interface {
	GetSocietyInfo() (*SocietyInfo, error)
}
