package models

import (
	"circledigital.in/real-state-erp/utils/custom"
	"github.com/google/uuid"
	"time"
)

// User model
type User struct {
	OrgId          uuid.UUID       `gorm:"not null;index" json:"orgId"`
	Organization   *Organization   `gorm:"foreignKey:OrgId;not null;constraint:OnUpdate:CASCADE" json:"organization,omitempty"`
	Name           string          `gorm:"not null" json:"name"`
	Email          string          `gorm:"primaryKey" json:"email"`
	Role           custom.UserRole `gorm:"not null" json:"role"`
	CreatedAt      time.Time       `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt      time.Time       `gorm:"autoUpdateTime" json:"updatedAt"`
	ProfilePicture string          `json:"profilePicture"`
}

func (u User) GetCreatedAt() time.Time {
	return u.CreatedAt
}
