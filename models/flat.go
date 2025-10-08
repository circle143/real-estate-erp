package models

import (
	"fmt"
	"log"
	"strconv"
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
	HeadingInstallment     = "Installment"
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
	ID      *uuid.UUID // used for payment plan only
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

func (f Flat) GetRowData(headers []Header, towerName string, print SafePrint, activeTowerPaymentPlans []TowerPaymentStatus) []string {
	var row []string

	totalPaidRemaining := decimal.Zero
	totalPayableAmount := decimal.Zero
	if f.SaleDetail != nil {
		row = append(row, f.SaleDetail.SaleNumber)
		totalPaidRemaining = f.SaleDetail.PaidAmount()
		totalPayableAmount = f.SaleDetail.GetTotalPayableAmount()
	} else {
		row = append(row, "")
	}

	// Recursive helper to flatten headers and fetch values
	var fill func(hs []Header, parent string, level int, parentID *uuid.UUID)
	fill = func(hs []Header, parent string, level int, parentID *uuid.UUID) {
		for _, h := range hs {
			// if leaf or level is 1
			if level == 1 || len(h.Items) == 0 {
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
				} else if f.SaleDetail == nil || f.SaleDetail.PaymentPlanRatio == nil {
					row = append(row, "")
					continue
				} else if strings.HasPrefix(parent, HeadingSale) {
					// sale
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
				} else if strings.HasPrefix(parent, HeadingBroker) {
					// broker
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
				} else if strings.HasPrefix(parent, HeadingPaymentPlan) {
					// payment plan info
					switch h.Heading {
					case "Name":
						row = append(row, f.SaleDetail.PaymentPlanRatio.PaymentPlanGroup.Name)
					case "Ratio":
						row = append(row, f.SaleDetail.PaymentPlanRatio.Ratio)
					default:
						row = append(row, "")
					}
				} else if strings.HasPrefix(parent, HeadingCustomer) {
					// customer
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
				} else if strings.HasPrefix(parent, HeadingCompanyCustomer) {
					// Company Customer
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
				} else if strings.HasPrefix(parent, HeadingPricebreakdown) {
					// price breakdown
					row = append(row, f.SaleDetail.PriceBreakdown.GetPriceFromSummary(h.Heading).String())
				} else if strings.HasPrefix(parent, HeadingInstallment) {
					// installment handling
					receipts := f.SaleDetail.Receipts

					// curent heading will be installment number, installment has 4 sub heading
					ind, convErr := strconv.Atoi(h.Heading)
					if convErr != nil ||
						// check if receipt is present
						len(receipts) < ind {
						row = append(row, make([]string, 6)...)
						continue
					}

					requiredReceipt := receipts[ind-1]
					row = append(row, requiredReceipt.ReceiptNumber)
					row = append(row, formatDateTime(requiredReceipt.CreatedAt))
					row = append(row, requiredReceipt.TotalAmount.String())
					row = append(row, string(requiredReceipt.Mode))
					row = append(row, requiredReceipt.GetReceiptStatus())
					if requiredReceipt.Cleared != nil {
						row = append(row, formatDateTime(requiredReceipt.Cleared.CreatedAt))
					} else {
						row = append(row, "")
					}
				} else {
					// this is payment plan row
					if parentID != nil && f.SaleDetail.PaymentPlanRatioId == *parentID {
						// handle payment plan here
						financeDetail, collectionDate := f.SaleDetail.PaymentPlanRatio.GetRatioAmountDetail(*h.ID, totalPayableAmount, totalPaidRemaining, f.ActivePaymentPlanRatioItems, activeTowerPaymentPlans)

						if financeDetail != nil {
							row = append(row, formatDateTime(*collectionDate))
							row = append(row, financeDetail.Total.String())
							row = append(row, financeDetail.Paid.String())
							row = append(row, financeDetail.Remaining.String())

							totalPaidRemaining = totalPaidRemaining.Sub(financeDetail.Paid)
							continue
						}
					}
					count := len(h.Items)
					row = append(row, make([]string, max(1, count))...)

				}
			} else if len(h.Items) > 0 {
				// Recursive for nested headers (Payment Plans, Installments, etc.)
				nextParent := parent
				if parent == "" {
					nextParent = h.Heading
				} else {
					nextParent = fmt.Sprintf("%s-%s", parent, h.Heading)
				}
				fill(h.Items, nextParent, level+1, h.ID)
			}
		}
	}

	fill(headers, "", 0, nil)
	for i, v := range row {
		if strings.TrimSpace(v) == "" {
			row[i] = "-"
		}
	}

	return row
}

func formatDateTime(t time.Time) string {
	return t.Format("02-01-2006")
}
