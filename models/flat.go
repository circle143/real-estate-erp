package models

import (
	"fmt"
	"log"
	"strings"
	"time"

	"circledigital.in/real-state-erp/utils/custom"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

const (
	HeadingFlat            = "Flat Details"
	HeadingSale            = "Sale Details"
	HeadingBroker          = "Broker Details"
	HeadingPaymentPlan     = "Payment Plan"
	HeadingCustomer        = "Customer Details"
	HeadingCompanyCustomer = "Company Customer Details"
	HeadingPricebreakdown  = "Price Breakdown"
)

// Flat model
// todo method to generate flat name from floor number, flat count in floor, and tower name
type Flat struct {
	Id      uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TowerId uuid.UUID `gorm:"not null;index;uniqueIndex:tower_flat_unique" json:"towerId"`
	Tower   *Tower    `gorm:"foreignKey:TowerId;not null" json:"tower,omitempty"`
	//FlatTypeId  uuid.UUID       `gorm:"not null;index" json:"flatTypeId"`
	//FlatType    *FlatType       `gorm:"foreignKey:FlatTypeId;not null" json:"flatType,omitempty"`
	Name                        string              `gorm:"not null;uniqueIndex:tower_flat_unique" json:"name"`
	FloorNumber                 int                 `gorm:"not null" json:"floorNumber"`
	Facing                      custom.Facing       `gorm:"not null;default:Default" json:"facing"`
	SaleableArea                decimal.Decimal     `gorm:"not null;type:numeric" json:"salableArea"`
	UnitType                    string              `gorm:"not null" json:"unitType"`
	CreatedAt                   time.Time           `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt                   time.Time           `gorm:"autoUpdateTime" json:"updatedAt"`
	SaleDetail                  *Sale               `gorm:"foreignKey:FlatId" json:"saleDetail,omitempty"`
	ActivePaymentPlanRatioItems []FlatPaymentStatus `gorm:"foreignKey:FlatId" json:"-"`
	//DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type Header struct {
	Heading string
	Items   []Header
}

func (f Flat) GetCreatedAt() time.Time {
	return f.CreatedAt
}

type SafePrint struct {
	ShouldPrint bool
}

func (p SafePrint) Print(s any) {
	if p.ShouldPrint {
		log.Println(s)
	}
}

func (f Flat) GetRowData(headers []Header, towerName string, print SafePrint) []string {
	var row []string

	// Recursive helper to flatten headers and fetch values
	var fill func(hs []Header, parent string)
	fill = func(hs []Header, parent string) {
		for _, h := range hs {
			if len(h.Items) == 0 {
				// Flat
				if strings.HasPrefix(parent, HeadingFlat) {
					switch h.Heading {
					case "Flat":
						row = append(row, f.Name)
					case "Floor":
						row = append(row, fmt.Sprintf("%d", f.FloorNumber))
					case "Facing":
						row = append(row, string(f.Facing))
					case "Saleable Area":
						row = append(row, f.SaleableArea.String())
					case "Unit Type":
						row = append(row, f.UnitType)
					case "Tower":
						row = append(row, towerName)
					default:
						row = append(row, "")
					}
					continue
				}

				if f.SaleDetail == nil {
					row = append(row, "")
					continue
				}

				// Sale
				if strings.HasPrefix(parent, HeadingSale) {
					switch h.Heading {
					case "ID":
						row = append(row, f.SaleDetail.SaleNumber)
					case "Total Price":
						row = append(row, f.SaleDetail.TotalPrice.String())
					case "Total Payable Amount":
						row = append(row, f.SaleDetail.GetTotalPayableAmount().String())
					case "Paid Amount":
						row = append(row, f.SaleDetail.PaidAmount().String())
					case "Pending Amount":
						row = append(row, f.SaleDetail.Pending().String())

					default:
						row = append(row, "")
					}
				}

				// Broker
				if strings.HasPrefix(parent, HeadingBroker) {
					switch h.Heading {
					case "Name":
						row = append(row, f.SaleDetail.Broker.Name)
					case "Aadhar":
						row = append(row, f.SaleDetail.Broker.AadharNumber)
					case "PAN":
						row = append(row, f.SaleDetail.Broker.PanNumber)
					default:
						row = append(row, "")
					}
				}

				// Payment Plan Info
				if strings.HasPrefix(parent, HeadingPaymentPlan) {
					switch h.Heading {
					case "Name":
						row = append(row, f.SaleDetail.PaymentPlanRatio.PaymentPlanGroup.Name)
					case "Ratio":
						row = append(row, f.SaleDetail.PaymentPlanRatio.Ratio)
					default:
						row = append(row, "")
					}
				}

				// Customer
				if strings.HasPrefix(parent, HeadingCustomer) {
					if len(f.SaleDetail.Customers) == 0 {
						row = append(row, "")
						continue
					}
					customer := f.SaleDetail.Customers[0]
					switch h.Heading {
					case "Name":
						row = append(row, customer.FirstName+" "+customer.LastName)
					case "Gender":
						row = append(row, string(customer.Gender))
					case "Email":
						row = append(row, customer.Email)
					case "Phone Number":
						row = append(row, customer.PhoneNumber)
					case "Nationality":
						row = append(row, string(customer.Nationality))
					case "Aadhar":
						row = append(row, customer.AadharNumber)
					case "PAN":
						row = append(row, customer.PanNumber)
					case "Passport Number":
						row = append(row, customer.PassportNumber)
					case "Profession":
						row = append(row, customer.Profession)
					case "Company Name":
						row = append(row, customer.CompanyName)
					default:
						row = append(row, "")
					}
				}

				// Company Customer
				if strings.HasPrefix(parent, HeadingCompanyCustomer) {
					if f.SaleDetail.CompanyCustomer == nil {
						row = append(row, "")
						continue
					}

					companyCustomer := f.SaleDetail.CompanyCustomer

					switch h.Heading {
					case "Name":
						row = append(row, companyCustomer.Name)
					case "Company PAN":
						row = append(row, companyCustomer.CompanyPan)
					case "GST":
						row = append(row, companyCustomer.CompanyGst)
					case "Aadhar":
						row = append(row, companyCustomer.AadharNumber)
					case "PAN":
						row = append(row, companyCustomer.PanNumber)
					default:
						row = append(row, "")
					}
				}

				// price breakdown
				if strings.HasPrefix(parent, HeadingPricebreakdown) {
					row = append(row, f.SaleDetail.PriceBreakdown.GetPriceFromSummary(h.Heading).String())
				}
			} else {
				// Recursive for nested headers (Payment Plans, Installments, etc.)
				nextParent := parent
				if parent == "" {
					nextParent = h.Heading
				} else {
					nextParent = fmt.Sprintf("%s-%s", parent, h.Heading)
				}
				fill(h.Items, nextParent)
			}
		}
	}

	fill(headers, "")
	for i, v := range row {
		if strings.TrimSpace(v) == "" {
			row[i] = "-"
		}
	}

	print.Print(row)
	return row
}
