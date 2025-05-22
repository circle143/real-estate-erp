package models

import (
	"circledigital.in/real-state-erp/utils/custom"
	"github.com/google/uuid"
	"time"
)

// Customer model
type Customer struct {
	Id               uuid.UUID            `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	SaleId           uuid.UUID            `gorm:"not null;index" json:"saleId"`
	Salutation       custom.Salutation    `gorm:"not null" json:"salutation"`
	FirstName        string               `gorm:"not null" json:"firstName"`
	LastName         string               `gorm:"not null" json:"lastName"`
	DateOfBirth      custom.DateOnly      `gorm:"type:date;not null" json:"dateOfBirth"`
	Gender           custom.Gender        `gorm:"not null" json:"gender"`
	Photo            string               `json:"photo"`
	MaritalStatus    custom.MaritalStatus `gorm:"not null" json:"maritalStatus"`
	Nationality      custom.Nationality   `gorm:"not null" json:"nationality"`
	Email            string               `gorm:"not null" json:"email"`
	PhoneNumber      string               `gorm:"not null" json:"phoneNumber"`
	MiddleName       string               `json:"middleName"`
	NumberOfChildren int                  `json:"numberOfChildren"`
	AnniversaryDate  *custom.DateOnly     `gorm:"type:date" json:"anniversaryDate"`
	AadharNumber     string               `json:"aadharNumber"`
	PanNumber        string               `json:"panNumber"`
	PassportNumber   string               `json:"passportNumber"`
	Profession       string               `json:"profession"`
	Designation      string               `json:"designation"`
	CompanyName      string               `json:"companyName"`
	CreatedAt        time.Time            `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt        time.Time            `gorm:"autoUpdateTime" json:"updatedAt"`
	//DeletedAt        gorm.DeletedAt       `gorm:"index"`
}

func (u Customer) GetCreatedAt() time.Time {
	return u.CreatedAt
}
