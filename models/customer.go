package models

import (
	"circledigital.in/real-state-erp/utils/custom"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"time"
)

// Customer model
type Customer struct {
	Id     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid();constraint:OnUpdate:CASCADE" json:"id"`
	FlatId uuid.UUID `gorm:"not null;index" json:"flatId"`
	//Flat             *Flat                `gorm:"foreignKey:FlatId;not null" json:"flat,omitempty"`
	Level            int                  `gorm:"not null" json:"level"`
	Salutation       custom.Salutation    `gorm:"not null" json:"salutation"`
	FirstName        string               `gorm:"not null" json:"firstName"`
	LastName         string               `gorm:"not null" json:"lastName"`
	DateOfBirth      time.Time            `gorm:"not null" json:"dateOfBirth"`
	Gender           custom.Gender        `gorm:"not null" json:"gender"`
	Photo            string               `gorm:"not null" json:"photo"`
	MaritalStatus    custom.MaritalStatus `gorm:"not null" json:"maritalStatus"`
	Nationality      custom.Nationality   `gorm:"not null" json:"nationality"`
	Email            pq.StringArray       `gorm:"type:text[];not null" json:"email"`
	PhoneNumber      pq.StringArray       `gorm:"type:text[];not null" json:"phoneNumber"`
	MiddleName       string               `json:"middleName"`
	NumberOfChildren int                  `json:"numberOfChildren"`
	AnniversaryDate  time.Time            `json:"anniversaryDate"`
	AadharNumber     string               `json:"aadharNumber"`
	PanNumber        string               `json:"panNumber"`
	PassportNumber   string               `json:"passportNumber"`
	Profession       string               `json:"profession"`
	Designation      string               `json:"designation"`
	CompanyName      string               `json:"companyName"`
	CreatedAt        time.Time            `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time            `gorm:"autoUpdateTime" json:"updatedAt"`
}

func (u Customer) GetCreatedAt() time.Time {
	return u.CreatedAt
}