package reports

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	"circledigital.in/real-state-erp/models"
	"circledigital.in/real-state-erp/utils/custom"
	"circledigital.in/real-state-erp/utils/payload"
	"github.com/go-chi/chi/v5"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func newMasterReportSheet(file *excelize.File, tower models.Tower) error {
	sheetName := tower.Name
	_, err := file.NewSheet(sheetName)
	if err != nil {
		return err
	}

	// Header row (UUID + time fields removed)
	headers := []string{
		"Tower_Name", "Flat_Name", "FloorNumber", "Facing", "SaleableArea", "UnitType",
		"SaleNumber", "TotalPrice",
		"BrokerName", "BrokerAadhar", "BrokerPan",
		"PaymentPlanGroup", "PaymentPlanRatio",
		"CustomerName", "CustomerEmail", "CustomerPhone", "CustomerPan", "CustomerAadhar",
		"CompanyName", "CompanyPan", "CompanyGst",
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if err := file.SetCellValue(sheetName, cell, h); err != nil {
			return err
		}
	}

	// Data rows
	rowIdx := 2
	for _, flat := range tower.Flats {
		baseFlat := []any{
			sheetName,
			flat.Name,
			flat.FloorNumber,
			flat.Facing,
			flat.SaleableArea.String(),
			flat.UnitType,
		}

		if flat.SaleDetail != nil {
			sale := flat.SaleDetail

			// sale values (removed ID + timestamps)
			saleVals := []any{
				sale.SaleNumber,
				sale.TotalPrice.String(),
			}

			// broker values (removed ID)
			var brokerVals []any
			if sale.Broker != nil {
				brokerVals = []any{
					sale.Broker.Name,
					sale.Broker.AadharNumber,
					sale.Broker.PanNumber,
				}
			} else {
				brokerVals = []any{"-", "-", "-"}
			}

			// payment plan group + ratio
			var planVals []any
			if sale.PaymentPlanRatio != nil && sale.PaymentPlanRatio.PaymentPlanGroup != nil {
				planVals = []any{
					sale.PaymentPlanRatio.PaymentPlanGroup.Name,
					sale.PaymentPlanRatio.Ratio,
				}
			} else {
				planVals = []any{"-", "-"}
			}

			// customers
			if len(sale.Customers) > 0 {
				for _, cust := range sale.Customers {
					custVals := []any{
						fmt.Sprintf("%s %s %s", cust.FirstName, cust.MiddleName, cust.LastName),
						cust.Email,
						cust.PhoneNumber,
						cust.PanNumber,
						cust.AadharNumber,
					}
					companyVals := []any{"-", "-", "-"}
					rowVals := append(append(append(append(baseFlat, saleVals...), brokerVals...), planVals...), custVals...)
					rowVals = append(rowVals, companyVals...)
					for colIdx, val := range rowVals {
						cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx)
						if err := file.SetCellValue(sheetName, cell, val); err != nil {
							return err
						}
					}
					rowIdx++
				}
			} else if sale.CompanyCustomer != nil {
				// company customer row (removed ID)
				custVals := []any{"-", "-", "-", "-", "-"}
				companyVals := []any{
					sale.CompanyCustomer.Name,
					sale.CompanyCustomer.CompanyPan,
					sale.CompanyCustomer.CompanyGst,
				}
				rowVals := append(append(append(append(baseFlat, saleVals...), brokerVals...), planVals...), custVals...)
				rowVals = append(rowVals, companyVals...)
				for colIdx, val := range rowVals {
					cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx)
					if err := file.SetCellValue(sheetName, cell, val); err != nil {
						return err
					}
				}
				rowIdx++
			} else {
				// sale without customer
				custVals := []any{"-", "-", "-", "-", "-"}
				companyVals := []any{"-", "-", "-"}
				rowVals := append(append(append(append(baseFlat, saleVals...), brokerVals...), planVals...), custVals...)
				rowVals = append(rowVals, companyVals...)
				for colIdx, val := range rowVals {
					cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx)
					if err := file.SetCellValue(sheetName, cell, val); err != nil {
						return err
					}
				}
				rowIdx++
			}
		} else {
			// flat without sale
			emptySale := []any{"-", "-"}
			emptyBroker := []any{"-", "-", "-"}
			emptyPlan := []any{"-", "-"}
			emptyCust := []any{"-", "-", "-", "-", "-"}
			emptyCompany := []any{"-", "-", "-"}
			rowVals := append(append(append(append(baseFlat, emptySale...), emptyBroker...), emptyPlan...), emptyCust...)
			rowVals = append(rowVals, emptyCompany...)
			for colIdx, val := range rowVals {
				cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx)
				if err := file.SetCellValue(sheetName, cell, val); err != nil {
					return err
				}
			}
			rowIdx++
		}
	}

	return nil
}

func generateMasterReport(db *gorm.DB, orgId, society string) (*bytes.Buffer, error) {
	var towerData []models.Tower
	err := db.
		Where("org_id = ? AND society_id = ?", orgId, society).
		Preload("Flats").
		Preload("Flats.SaleDetail").
		Preload("Flats.SaleDetail.PaymentPlanRatio").
		Preload("Flats.SaleDetail.PaymentPlanRatio.PaymentPlanGroup").
		Preload("Flats.SaleDetail.Receipts").
		Preload("Flats.SaleDetail.Receipts.Cleared").
		Preload("Flats.SaleDetail.Broker").
		Preload("Flats.SaleDetail.Customers").
		Preload("Flats.SaleDetail.CompanyCustomer").
		Find(&towerData).Error
	if err != nil {
		return nil, err
	}

	if len(towerData) < 1 {
		return nil, &custom.RequestError{
			Status:  http.StatusNoContent,
			Message: "No tower found",
		}
	}

	var reportFile *excelize.File = nil
	if os.Getenv("ENV") == "development" {
		reportFile = excelize.NewFile()
	} else {
		reportFile = excelize.NewFile(
			excelize.Options{
				Password: society,
			},
		)
	}

	for _, tower := range towerData {
		sheetErr := newMasterReportSheet(reportFile, tower)
		if sheetErr != nil {
			return nil, sheetErr
		}
	}

	// Delete the default "Sheet1"
	if err := reportFile.DeleteSheet("Sheet1"); err != nil {
		return nil, err
	}

	// Write Excel file to buffer
	var buf bytes.Buffer
	if err := reportFile.Write(&buf); err != nil {
		return nil, err
	}
	return &buf, nil
}

// generateMasterReport() generates a master report
// master report contains all the flats and its related sale details
func (s *reportService) generateMasterReport(w http.ResponseWriter, r *http.Request) {
	// get society rera and org id from request
	orgId := r.Context().Value(custom.OrganizationIDKey).(string)
	societyRera := chi.URLParam(r, "society")

	report, err := generateMasterReport(s.db, orgId, societyRera)
	if err != nil {
		payload.HandleError(w, err)
		return
	}

	// Set headers so browser/download tools recognize it as Excel
	w.Header().Set("Content-Type",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_master_report.xlsx", societyRera))
	w.Header().Set("Content-Length", fmt.Sprint(report.Len()))

	// Write to response
	if _, err := w.Write(report.Bytes()); err != nil {
		payload.HandleError(w, err)
		return
	}
}
